package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/asgaines/blockchain/chain"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"google.golang.org/grpc"
)

// BlockchainFile is a simple storage of the blockchain in a file
const BlockchainFile = "storage.json"

// InitialExpectedHashrate is a guess of how many hashes are possible per second.
// It seeds the network for the first block, but replaced by real data once it
// comes through.
//
// Do not set it unreasonably high, or the first block could never be solved.
const InitialExpectedHashrate = float64(1_6) //00_000)

// Node represents a blockchain client
// It preserves a copy of the blockchain, competes for new block additions by mining,
// and verifies work of peer nodes
type Node interface {
	Run(ctx context.Context)
	Mine()
	GetChain() chain.Chain
	DiscoverPeers() []Peer
	PropagateTransaction(tx *pb.Tx)
	Close()

	pb.BlockchainServer
}

// NewNode instantiates a Node; a blockchain client
func NewNode(id int32, ID string, targetDur time.Duration, recalcPeriod int) Node {
	n := node{
		id:           id,
		ID:           ID,
		queue:        make([]*pb.Tx, 0),
		targetDur:    targetDur,
		hashesPerSec: InitialExpectedHashrate,
		startTime:    time.Now(),
		recalcPeriod: recalcPeriod,
	}

	n.chain = n.initChain()
	n.target = n.calcTarget()
	n.peers = n.DiscoverPeers()

	return &n
}

type node struct {
	id           int32
	ID           string
	peers        []Peer
	chain        chain.Chain
	queue        []*pb.Tx
	nonce        uint64
	startTime    time.Time
	numHashes    uint64
	hashesPerSec float64
	target       uint64
	targetDur    time.Duration
	rates        []float64
	recalcPeriod int
}

func (n node) initChain() chain.Chain {
	f, err := os.Open(n.getStorageFname())
	if err != nil {
		log.Println("Initializing node with new blockchain")
		return chain.NewChain()
	}
	defer f.Close()

	var bc chain.Chain

	decoder := json.NewDecoder(f)
	if err = decoder.Decode(&bc); err != nil {
		log.Fatal(err)
	}

	if !bc.IsSolid() {
		log.Fatalf("Initialization failed due to broken chain in storage file: %s", BlockchainFile)
	}

	log.Println("Initializing node with blockchain from storage")

	return bc
}

func (n *node) Run(ctx context.Context) {
	defer n.storeChain()
	defer n.Close()
	go n.Mine()
	<-ctx.Done()
}

func (n *node) AddTxToQueue(tx *pb.Tx) {
	log.Printf("Received new tx: %s", tx)
	n.queue = append(n.queue, tx)
	n.PropagateTransaction(tx)
}

func (n *node) Mine() {
	for {
		block := chain.NewBlock(
			n.chain[len(n.chain)-1],
			n.queue,
			n.nonce,
			n.target,
			n.ID,
		)

		n.numHashes++

		if solved := block.Hash <= n.target; solved {
			log.Printf("%064b\n", block.Target)
			log.Printf("%064b\n", block.Hash)

			log.Printf("Took %v\n", n.chain.TimeSinceLastLink())

			ave := time.Since(n.startTime).Seconds() / float64(len(n.chain))
			log.Printf("Average %v seconds/block\n", ave)
			n.rates = append(n.rates, ave)

			log.Println()

			chain := n.chain.AddBlock(block)
			n.setChain(chain)
		}

		n.nonce++

		// Use this to control the amount of hashes / CPU usage
		time.Sleep(10 * time.Millisecond)
	}
}

func (n *node) GetChain() chain.Chain {
	return n.chain
}

func (n *node) setChain(chain chain.Chain) {
	if chain.IsSolid() && len(chain) > len(n.chain) {
		log.Printf("Overriding with new chain: %v\n", n.chain)

		n.chain = chain

		if len(chain)%n.recalcPeriod == 0 {
			n.updateTarget()
		}

		// TODO: ensure ALL txs in queue are in new chain
		// If not, keep orphans to queue
		n.queue = make([]*pb.Tx, 0)

		n.propagate()
	} else {
		log.Printf("%d: Not overriding\n", n.id)
	}
}

func (n *node) DiscoverPeers() []Peer {
	doors := []string{
		"127.0.0.1:5050",
		"127.0.0.1:5051",
		"127.0.0.1:5052",
		"127.0.0.1:5053",
		"127.0.0.1:5054",
		"127.0.0.1:5055",
		"127.0.0.1:5056",
		"127.0.0.1:5057",
		"127.0.0.1:5058",
		"127.0.0.1:5059",
	}

	ctx := context.Background()

	peers := make([]Peer, 0)

	var wg sync.WaitGroup
	wg.Add(len(doors))

	for _, door := range doors {
		go func(door string) {
			defer wg.Done()
			conn, err := grpc.Dial(door, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("could not dial: %s", err)
			}

			client := pb.NewBlockchainClient(conn)

			resp, err := client.Ping(ctx, &pb.PingRequest{Id: n.id})
			if err != nil {
				return
			}
			if resp.GetOk() {
				peers = append(peers, NewPeer(
					resp.GetId(),
					client,
					conn,
				))
			}
		}(door)
	}

	wg.Wait()

	log.Printf("Discovered peers: %s\n", peers)

	return peers
}

// updateTarget calculates a new mining target (block hash must be less than target
// to be considered solved), and distributes it to the network
func (n *node) updateTarget() {
	n.UpdateHashrate()
	n.target = n.calcTarget()
}

func (n node) PropagateTransaction(tx *pb.Tx) {
	for _, p := range n.peers {
		if err := p.SubmitTx(tx); err != nil {
			log.Printf("could not propagate tx: %s\n", err)
		}
	}
}

func (n *node) UpdateHashrate() {
	// This is overall average. Likely wanting a more instantaneous rate to
	// respond to changes more rapidly
	n.hashesPerSec = float64(n.numHashes) / time.Since(n.startTime).Seconds()
}

// calcTarget retrieves a target which must be greater than a hash value
// for a block to be considered solved.
// It is a determinant for how much work is expected for a
// new block to be added to chain.
func (n *node) calcTarget() uint64 {
	// Lowest possible difficulty (highest possible target)
	var difficulty1Target uint64 = 0xFF_FF_FF_FF_FF_FF_FF_FF

	shifts := math.Log2(n.hashesPerSec * n.targetDur.Seconds())
	target := difficulty1Target >> int(math.Round(shifts))

	return target
}

func (n node) propagate() {
	for _, p := range n.peers {
		log.Printf("%d: Propagation to %d\n\n", n.id, p.GetID())
		// TODO: propagate the chain across the network; claim your prize
		// if err := p.SubmitChain(n.chain); err != nil {
		// 	log.Println(err)
		// }
	}
}

func (n *node) storeChain() error {
	f, err := os.OpenFile(n.getStorageFname(), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(n.chain.ToJSON()); err != nil {
		return err
	}

	return nil
}

func (n node) randomPeer() Peer {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	return n.peers[r.Intn(len(n.peers))]
}

func (n *node) Close() {
	for _, peer := range n.peers {
		if err := peer.Close(); err != nil {
			log.Println(err)
		}
	}

	log.Println("Shutting down node")
}

func (n *node) getStorageFname() string {
	return fmt.Sprintf("%s_%s", n.ID, BlockchainFile)
}
