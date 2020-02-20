package nodes

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/asgaines/blockchain/chain"
	"github.com/asgaines/blockchain/dmaps"
	"github.com/asgaines/blockchain/mining"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
)

// InitialExpectedHashrate is the seed of how many hashes are possible per second.
// The variable set by it is overridden by real data once it comes through.
//
// Setting it too high could lead to the genesis block solve taking a long time
// before the difficulty is adjusted.
const InitialExpectedHashrate = float64(50) // ultra: 700_000)

// Node represents a blockchain node; a peer within the network.
// It preserves a copy of the blockchain, competes for new block additions by mining,
// and verifies work of peer nodes.
type Node interface {
	Run(ctx context.Context)
	Ready() chan struct{}

	pb.NodeServer
}

// NewNode instantiates a Node; a blockchain client/peer for mining
// and propagating new blocks/transactions
func NewNode(miners []mining.Miner, pubkey string, poolID int, minPeers int, maxPeers int, targetDurPerBlock time.Duration, recalcPeriod int, returnAddr string, seedAddrs []string, speed mining.HashSpeed, filesPrefix string, hasher chain.Hasher) Node {
	n := node{
		miners:            miners,
		pubkey:            pubkey,
		poolID:            poolID,
		txpool:            make([]*pb.Tx, 0),
		peers:             make(map[NodeID]Peer),
		knownAddrs:        dmaps.New(),
		minPeers:          minPeers,
		maxPeers:          maxPeers,
		targetDurPerBlock: targetDurPerBlock,
		recalcPeriod:      recalcPeriod,
		returnAddr:        returnAddr,
		filesPrefix:       filesPrefix,
		hasher:            hasher,
		seedAddrs:         seedAddrs,
		ready:             make(chan struct{}),
	}

	n.appendAddrs(n.getSeedAddrs())

	f, err := os.OpenFile("/storage/"+filesPrefix+"_blocks.tsv", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}

	f2, err := os.OpenFile("/storage/"+filesPrefix+"_periods.tsv", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}

	n.dursF = f
	n.statsF = f2

	return &n
}

type node struct {
	// pubkey is the public key for the node's miner.
	// The rewards for mining a block by this node will be attributed to this
	// public key. The miner can run multiple nodes with the same pubkey by
	// modifying the poolID
	pubkey string
	// poolID allows a single pubkey to be used across multiple nodes. Each
	// node within the single miner's pool should have a unique ID.
	poolID            int
	miners            []mining.Miner
	txpool            []*pb.Tx
	peers             map[NodeID]Peer
	knownAddrs        dmaps.Dmap
	minPeers          int
	maxPeers          int
	chain             *chain.Chain
	targetDurPerBlock time.Duration
	recalcPeriod      int
	returnAddr        string
	dursF             *os.File
	statsF            *os.File
	filesPrefix       string
	difficulty        float64
	hasher            chain.Hasher
	seedAddrs         []string
	ready             chan struct{}
}

type nodeID struct {
	pubkey string
	id     int
}

func (n *node) Run(ctx context.Context) {
	defer func() {
		if err := n.chain.Store(n.filesPrefix); err != nil {
			log.Println(err)
		}
	}()
	defer n.close()

	log.Println("Discovering peers...")
	n.discoverPeers(ctx)

	log.Println("Fetching initial state from peers...")
	c, diff, err := n.getInitState(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if !n.IsValid(c) {
		c = chain.NewChain(n.hasher)
	}

	n.difficulty = diff
	n.chain = c

	n.resetTxpool()

	log.Println("Initializing mining...")
	prevHash := n.hasher.Hash(n.chain.LastLink())
	for _, miner := range n.miners {
		miner.SetTarget(n.difficulty)
		miner.UpdatePrevHash(prevHash)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		go n.periodicDiscoverPeers(ctx)
	}()

	log.Println("Mining started...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		n.mine(ctx)
	}()

	// Allow the server to begin accepting connections
	n.ready <- struct{}{}

	wg.Wait()
}

func (n *node) Ready() chan struct{} {
	return n.ready
}

func (n *node) close() {
	for _, peer := range n.peers {
		if err := peer.Close(); err != nil {
			log.Println(err)
		}
	}

	if err := n.dursF.Close(); err != nil {
		log.Println(err)
	}

	if err := n.statsF.Close(); err != nil {
		log.Println(err)
	}
}

func (n node) propagateChain(except map[NodeID]bool) {
	for nodeID, p := range n.peers {
		if _, ok := except[nodeID]; !ok {
			if err := p.ShareChain(n.chain, n.getID()); err != nil {
				log.Printf("Removing peer: %s", nodeID.Pubkey)
				delete(n.peers, nodeID)
			}
		}
	}
}

func (n node) propagateTx(tx *pb.Tx, except NodeID) {
	for nodeID, p := range n.peers {
		if nodeID != except {
			if err := p.ShareTx(tx, n.getID()); err != nil {
				// log.Println(err)
			}
		}
	}
}

func (n *node) getCreditFor(pubkey string) float64 {
	creditInChain := n.chain.GetCreditFor(pubkey)

	debitsInTxpool := float64(0)
	for _, tx := range n.txpool {
		if tx.GetSender() == pubkey {
			debitsInTxpool += tx.GetValue()
		}
	}

	return creditInChain - debitsInTxpool
}

func (n *node) getStorageFnameProto() string {
	return fmt.Sprintf("%s.proto", n.filesPrefix)
}

func (n *node) getStorageFnameJSON() string {
	return fmt.Sprintf("%s.json", n.filesPrefix)
}

type SubmitReport struct {
	chain chain.Chain
}

func (n *node) mergeSubmits(chans ...<-chan SubmitReport) <-chan SubmitReport {
	submissions := make(chan SubmitReport)

	go func() {
		var wg sync.WaitGroup

		wg.Add(len(chans))
		for _, c := range chans {
			go func(c <-chan SubmitReport) {
				for submission := range c {
					submissions <- submission
				}
				wg.Done()
			}(c)
		}

		wg.Wait()
		close(submissions)
	}()

	return submissions
}

func (n *node) getKnownAddrsExcept(except []string) []string {
	knownAddrs := n.knownAddrs.ReadAll()
	addrs := make([]string, 0, len(knownAddrs))

	exceptions := make(map[string]bool, len(except))
	for _, e := range except {
		exceptions[e] = true
	}

	for _, addr := range knownAddrs {
		if _, ok := exceptions[addr]; !ok {
			addrs = append(addrs, addr)
		}
	}

	return addrs
}

func (n *node) appendAddrs(addrs []string) {
	newAddrs := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		newAddrs = append(newAddrs, addr)
	}

	n.knownAddrs.WriteMany(newAddrs)
}

func (n *node) getSeedAddrs() []string {
	var addrs []string

	for addr := range n.getBootSeeds() {
		addrs = append(addrs, addr)
	}

	for _, addr := range n.seedAddrs {
		addrs = append(addrs, addr)
	}

	return addrs
}

func (n *node) getID() NodeID {
	return NodeID{
		Pubkey:     n.pubkey,
		Id:         int32(n.poolID),
		ReturnAddr: n.returnAddr,
	}
}
