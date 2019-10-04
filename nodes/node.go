package nodes

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/asgaines/blockchain/chain"
	"github.com/asgaines/blockchain/transactions"
)

// Node represents a blockchain client
// It preserves a copy of the blockchain, competes for new block additions by mining,
// and verifies work of peer nodes
type Node interface {
	Mine()
	GetID() int
	Run(ctx context.Context)
	SubmitTransaction(tx transactions.Tx)
	GetCC() chan chain.Chain
	GetChain() chain.Chain
	SetChain(bchain chain.Chain) error
	SetPeers(node []Node)
	UpdateTarget(target uint64)
	GetNumHashes() uint64
}

// SolveReport is a report of a node solving a block
type SolveReport struct {
	ID    int
	Block *chain.Block
}

// NewNode instantiates a Node; a blockchain client
func NewNode(id int, bchain chain.Chain, solveReport chan<- SolveReport, target uint64) Node {
	return &node{
		id:          id,
		chain:       bchain,
		solveReport: solveReport,
		cc:          make(chan chain.Chain),
		latency:     100 * time.Millisecond,
		queue:       make(transactions.Txs, 0),
		nonce:       rand.Uint64(),
		target:      target,
	}
}

type node struct {
	id          int
	peers       []Node
	chain       chain.Chain
	solveReport chan<- SolveReport
	cc          chan chain.Chain
	latency     time.Duration
	queue       transactions.Txs
	nonce       uint64
	numHashes   uint64
	target      uint64
}

func (n *node) Run(ctx context.Context) {
	go n.Mine()

	for {
		select {
		case chain := <-n.cc:
			if err := n.SetChain(chain); err != nil {
				log.Fatal(err)
			}
		case <-ctx.Done():
			log.Printf("%d: Shutting down\n", n.id)
			close(n.cc)
			close(n.solveReport)
			return
		}
	}
}

func (n *node) SubmitTransaction(tx transactions.Tx) {
	n.queue = append(n.queue, tx)
}

func (n *node) Mine() {
	for {
		block := chain.NewBlock(
			n.chain[len(n.chain)-1],
			n.queue,
			n.nonce,
			n.target,
		)
		n.numHashes++

		if solved := block.Hash <= n.target; solved {
			n.solveReport <- SolveReport{
				ID:    n.id,
				Block: block,
			}

			chain := n.chain.AddBlock(block)
			if err := n.SetChain(chain); err != nil {
				log.Fatal(err)
			}
		}

		n.nonce++
		time.Sleep(100 * time.Millisecond)
	}
}

func (n node) GetCC() chan chain.Chain {
	return n.cc
}

func (n node) GetID() int {
	return n.id
}

func (n *node) GetChain() chain.Chain {
	return n.chain
}

func (n *node) SetChain(chain chain.Chain) error {
	if chain.IsSolid() && len(chain) > len(n.chain) {
		// fmt.Printf("%d: Overriding with new chain\n", n.id)

		n.chain = chain

		// TODO: ensure ALL txs in queue are in new chain
		// If not, keep orphans to queue
		n.queue = make(transactions.Txs, 0)

		if err := n.store(); err != nil {
			return err
		}

		n.propagate()
	} else {
		// fmt.Printf("%d: Not overriding\n", n.id)
	}

	return nil
}

func (n *node) SetPeers(nodes []Node) {
	n.peers = nodes
}

func (n *node) UpdateTarget(target uint64) {
	n.target = target
}

func (n *node) GetNumHashes() uint64 {
	return n.numHashes
}

func (n node) propagate() {
	for _, peer := range n.peers {
		// fmt.Printf("%d: Propagation to %d\n\n", n.GetID(), peer.GetID())

		time.Sleep(n.latency)
		peer.GetCC() <- n.chain
	}
}

func (n node) store() error {
	fname := fmt.Sprintf("%s/%d_%s", "storage", n.GetID(), chain.BlockchainFile)

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
