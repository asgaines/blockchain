package nodes

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/asgaines/blockchain/chain"
	"github.com/asgaines/blockchain/dmaps"
	"github.com/asgaines/blockchain/mining"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

// InitialExpectedHashrate is the seed of how many hashes are possible per second.
// The variable set by it is overridden by real data once it comes through.
//
// Setting it too high could lead to the genesis block solve taking a long time
// before the difficulty is adjusted.
const InitialExpectedHashrate = float64(100)

// Node represents a blockchain node; a peer within the network.
// It preserves a copy of the blockchain, competes for new block additions by mining,
// and verifies work of peer nodes.
type Node interface {
	Run(ctx context.Context)

	pb.NodeServer
}

// NewNode instantiates a Node; a blockchain client for mining
// and propagating new blocks/transactions
func NewNode(pubkey string, poolID int, minPeers int, maxPeers int, targetDur time.Duration, recalcPeriod int, serverPort int, speed mining.HashSpeed, ratesFileName string) Node {
	n := node{
		pubkey:       pubkey,
		poolID:       poolID,
		peers:        make(map[NodeID]Peer),
		knownAddrs:   dmaps.New(),
		minPeers:     minPeers,
		maxPeers:     maxPeers,
		targetDur:    targetDur,
		recalcPeriod: recalcPeriod,
		serverPort:   serverPort,
	}

	n.chain = n.initChain()

	n.miner = mining.NewMiner(
		(*chain.Block)(n.chain.Blocks[n.chain.Length()-1]),
		pubkey,
		InitialExpectedHashrate*targetDur.Seconds(),
		targetDur,
		speed,
	)

	n.appendAddrs(n.getSeedAddrs())

	f, err := os.OpenFile(ratesFileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}

	n.ratesF = f

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
	poolID       int
	miner        mining.Miner
	peers        map[NodeID]Peer
	knownAddrs   dmaps.Dmap
	minPeers     int
	maxPeers     int
	chain        *chain.Chain
	startTime    time.Time
	targetDur    time.Duration
	recalcPeriod int
	serverPort   int
	ratesF       *os.File
}

type nodeID struct {
	pubkey string
	id     int
}

func (n *node) Run(ctx context.Context) {
	// defer func() {
	// 	if err := n.storeChain(); err != nil {
	// 		log.Println(err)
	// 	}
	// }()
	defer n.close()

	n.startTime = time.Now()

	// go n.periodicDiscoverPeers(ctx)
	go n.mine(ctx)

	<-ctx.Done()
}

func (n *node) close() {
	for _, peer := range n.peers {
		if err := peer.Close(); err != nil {
			log.Println(err)
		}
	}

	if err := n.ratesF.Close(); err != nil {
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

func (n *node) storeChain() error {
	f, err := os.OpenFile(n.getStorageFnameProto(), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}()

	b, err := proto.Marshal(n.chain.ToProto())
	if err != nil {
		return fmt.Errorf("could not marshal chain: %w", err)
	}

	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}

	fj, err := os.OpenFile(n.getStorageFnameJSON(), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if err := fj.Close(); err != nil {
			log.Println(err)
		}
	}()

	if err := new(jsonpb.Marshaler).Marshal(fj, n.chain.ToProto()); err != nil {
		return err
	}

	return nil
}

func (n *node) getStorageFnameProto() string {
	return fmt.Sprintf("%s.proto", n.pubkey)
}

func (n *node) getStorageFnameJSON() string {
	return fmt.Sprintf("%s.json", n.pubkey)
}

type SubmitReport struct {
	chain chain.Chain
}

func (n node) initChain() *chain.Chain {
	b, err := ioutil.ReadFile(n.getStorageFnameProto())
	if err != nil {
		return chain.NewChain()
	}

	var bcpb pb.Chain

	if err := proto.Unmarshal(b, &bcpb); err != nil {
		log.Fatalf("could not unmarshal chain: %s", err)
	}

	bc := chain.Chain(bcpb)

	if !bc.IsSolid() {
		log.Fatalf("Initialization failed due to broken chain in storage file: %s", n.getStorageFnameProto())
	}

	return &bc
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
