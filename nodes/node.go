package nodes

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/asgaines/blockchain/chain"
	"github.com/asgaines/blockchain/dmaps"
	"github.com/asgaines/blockchain/mining"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
)

// Node represents a blockchain node; a peer within the network.
// It preserves a copy of the blockchain, competes for new block additions by mining,
// and verifies work of peer nodes.
type Node interface {
	Run(ctx context.Context)

	pb.NodeServer
}

// NewNode instantiates a Node; a blockchain client for mining
// and propagating new blocks/transactions
func NewNode(c *chain.Chain, miners []mining.Miner, pubkey string, poolID int, minPeers int, maxPeers int, targetDurPerBlock time.Duration, recalcPeriod int, serverPort int, speed mining.HashSpeed, filesPrefix string, difficulty float64, hasher chain.Hasher) Node {
	n := node{
		chain:             c,
		miners:            miners,
		pubkey:            pubkey,
		poolID:            poolID,
		peers:             make(map[NodeID]Peer),
		knownAddrs:        dmaps.New(),
		minPeers:          minPeers,
		maxPeers:          maxPeers,
		targetDurPerBlock: targetDurPerBlock,
		recalcPeriod:      recalcPeriod,
		serverPort:        serverPort,
		filesPrefix:       filesPrefix,
		difficulty:        difficulty,
		hasher:            hasher,
	}

	// n.appendAddrs(n.getSeedAddrs())

	f, err := os.OpenFile(filesPrefix+"_blocks.tsv", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}

	f2, err := os.OpenFile(filesPrefix+"_periods.tsv", os.O_WRONLY|os.O_CREATE, 0644)
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
	peers             map[NodeID]Peer
	knownAddrs        dmaps.Dmap
	minPeers          int
	maxPeers          int
	chain             *chain.Chain
	targetDurPerBlock time.Duration
	recalcPeriod      int
	serverPort        int
	dursF             *os.File
	statsF            *os.File
	filesPrefix       string
	difficulty        float64
	hasher            chain.Hasher
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

	var wg sync.WaitGroup

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	go n.periodicDiscoverPeers(ctx)
	// }()

	wg.Add(1)
	go func() {
		defer wg.Done()
		n.mine(ctx)
	}()

	wg.Wait()
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

func (n node) propagateTx(tx *pb.Tx, except NodeID) {
	for nodeID, p := range n.peers {
		if nodeID != except {
			if err := p.ShareTx(tx, n.getID(), int32(n.serverPort)); err != nil {
				log.Println(err)
			}
		}
	}
}

func (n node) propagateChain() {
	for nodeID, p := range n.peers {
		log.Printf("Propagation to %v\n\n", nodeID)
		if err := p.ShareChain(n.chain, n.getID(), int32(n.serverPort)); err != nil {
			log.Println(err)
			delete(n.peers, nodeID)
		}
	}
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

	extraAddrs := os.Getenv("BLOCKCHAIN_ADDRS")
	for _, addr := range strings.Split(extraAddrs, ",") {
		addrs = append(addrs, addr)
	}

	return addrs
}

func (n *node) getID() NodeID {
	return NodeID{
		Pubkey: n.pubkey,
		Id:     int32(n.poolID),
	}
}
