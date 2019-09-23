package nodes

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/asgaines/blockchain/blockchain"
	"github.com/asgaines/blockchain/transactions"
)

// Node represents a blockchain client
// It preserves a copy of the blockchain, competes for new block additions by mining,
// and verifies work of peer nodes
type Node interface {
	Mine()
	GetID() int
	Run(context.Context)
	SubmitTransaction(tx transactions.Transaction)
	GetCC() chan blockchain.Blockchain
	GetChain() blockchain.Blockchain
	SetChain(blockchain.Blockchain) error
	SetPeers([]Node)
	SenseDifficulty() float64
}

// NewNode instantiates a Node; a blockchain client
func NewNode(id int, chain blockchain.Blockchain, rcvTx chan transactions.Transaction, targetMin float64) Node {
	return &node{
		id:        id,
		chain:     chain,
		rcvTx:     rcvTx,
		cc:        make(chan blockchain.Blockchain),
		latency:   100 * time.Millisecond,
		targetMin: targetMin,
		queue:     make(transactions.Transactions, 0),
	}
}

type node struct {
	id        int
	peers     []Node
	chain     blockchain.Blockchain
	rcvTx     chan transactions.Transaction
	cc        chan blockchain.Blockchain
	latency   time.Duration
	targetMin float64
	queue     transactions.Transactions
}

func (n *node) Run(ctx context.Context) {
	go n.Mine()

	for {
		select {
		case tx := <-n.rcvTx:
			n.queue = append(n.queue, tx)
		case chain := <-n.cc:
			fmt.Printf("%d: Received chain posting from peer %v\n", n.id, chain)
			if err := n.SetChain(chain); err != nil {
				log.Fatal(err)
			}
		case <-ctx.Done():
			log.Printf("%d: Shutting down\n", n.id)
			close(n.cc)
			return
		}
	}
}

func (n *node) SubmitTransaction(tx transactions.Transaction) {
	n.rcvTx <- tx
}

func (n *node) Mine() {
	found := make(chan *blockchain.Block)

	go n.mine(found)

	for {
		select {
		case minedBlock := <-found:
			chain := n.chain.AddBlock(minedBlock)
			if err := n.SetChain(chain); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func (n *node) mine(found chan<- *blockchain.Block) {
	difficulty := n.SenseDifficulty()

	done := make(chan struct{})
	defer close(done)

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				difficulty = n.SenseDifficulty()
			case <-done:
				return
			}
		}
	}()

	nonce := rand.Intn(1000)
	orig := nonce

	for {
		block := blockchain.NewBlock(n.chain[len(n.chain)-1], n.queue, nonce)

		atLeast := float64(3)
		numZeroes := atLeast + (difficulty * (64 - atLeast))

		match := true
		for i := 0; i < int(numZeroes); i++ {
			if block.Hash[i] != '0' {
				match = false
				break
			}
		}

		if match {
			fmt.Printf("Node %d mined new block of hash %s. %d nonce updates. difficulty: %v\n", n.id, block.Hash, nonce-orig, difficulty)
			found <- block
		}

		nonce++
	}
}

func (n node) GetCC() chan blockchain.Blockchain {
	return n.cc
}

func (n node) GetID() int {
	return n.id
}

func (n *node) GetChain() blockchain.Blockchain {
	return n.chain
}

func (n *node) SetChain(chain blockchain.Blockchain) error {
	if chain.IsSolid() && len(chain) > len(n.chain) {
		fmt.Printf("%d: Overriding with new chain\n", n.id)

		n.chain = chain

		// TODO: ensure ALL txs in queue are in new chain
		// If not, keep orphans to queue
		n.queue = make(transactions.Transactions, 0)

		if err := n.store(); err != nil {
			return err
		}

		n.propagate()
	} else {
		fmt.Printf("%d: Not overriding\n", n.id)
	}

	return nil
}

func (n *node) SetPeers(nodes []Node) {
	n.peers = nodes
}

// SenseDifficulty retrieves a value between 0 and 1 where 1 is maximum difficulty.
// Difficulty is used to determine how much work is expected for a new block to be added
// to chain.
func (n *node) SenseDifficulty() float64 {
	lapse := n.GetChain().TimeSinceLastLink()

	scaled := -(1/n.targetMin)*float64(float64(lapse)/float64(time.Minute)) + 1

	return math.Max(scaled, 0)
}

func (n node) propagate() {
	for _, peer := range n.peers {
		fmt.Printf("%d: Propagation to %d\n\n", n.GetID(), peer.GetID())

		time.Sleep(n.latency)

		peer.GetCC() <- n.chain
	}
}

func (n node) store() error {
	fname := fmt.Sprintf("%s/%d_%s", "storage", n.GetID(), blockchain.BlockchainFile)

	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(n.chain.ToJSON()); err != nil {
		return err
	}

	return nil
}
