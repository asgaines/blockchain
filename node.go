package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
)

// Node represents a blockchain client
// It preserves a copy of the blockchain, competes for new block additions by mining,
// and verifies work of peer nodes
type Node interface {
	GetID() int
	Run(context.Context)
	SubmitTransaction(lb LilBits, podium chan<- Node)
	Mine(lb LilBits)
	GetCC() chan Blockchain
	GetChain() *Blockchain
	SetChain(Blockchain) error
	SetPeers([]Node)
	SenseDifficulty() float64
}

// NewNode instantiates a Node; a blockchain client
func NewNode(id int, chain Blockchain, transactions chan LilBits, targetMin float64) Node {
	return &node{
		id:           id,
		chain:        chain,
		transactions: transactions,
		cc:           make(chan Blockchain),
		latency:      100 * time.Millisecond,
		ilost:        make(chan struct{}),
		targetMin:    targetMin,
	}
}

type node struct {
	id           int
	peers        []Node
	chain        Blockchain
	transactions chan LilBits
	podium       chan<- Node
	cc           chan Blockchain
	latency      time.Duration
	mining       bool
	ilost        chan struct{}
	targetMin    float64
}

func (n *node) Run(ctx context.Context) {
	for {
		select {
		case lilbits := <-n.transactions:
			go n.Mine(lilbits)
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

func (n *node) SubmitTransaction(lb LilBits, podium chan<- Node) {
	n.podium = podium
	n.transactions <- lb
}

func (n *node) Mine(lb LilBits) {
	found := make(chan *Block)

	defer func() {
		n.mining = false
	}()
	n.mining = true

	go n.mine(lb, found)

	select {
	case block := <-found:
		n.mining = false
		n.podium <- n

		chain := n.chain.AddBlock(block)
		if err := n.SetChain(chain); err != nil {
			log.Fatal(err)
		}

		return
	case <-n.ilost:
		// A more robust blockchain is always mining, not only when a transaction
		// has been logged
		return
	}
}

func (n *node) mine(lb LilBits, found chan<- *Block) {
	difficulty := n.SenseDifficulty()

	done := make(chan struct{})
	defer close(done)

	go func() {
		ticker := time.NewTicker(1 * time.Second)
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
		block := NewBlock(n.chain[len(n.chain)-1], lb, nonce)

		numZeroes := 3 + (difficulty * (64 - 3))

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
			return
		}

		nonce++
	}
}

func (n node) GetCC() chan Blockchain {
	return n.cc
}

func (n node) GetID() int {
	return n.id
}

func (n *node) GetChain() *Blockchain {
	return &n.chain
}

func (n *node) SetChain(chain Blockchain) error {
	if chain.IsSolid() && len(chain) > len(n.chain) {
		fmt.Printf("%d: Overriding with new chain\n", n.id)
		n.chain = chain

		if n.mining {
			n.ilost <- struct{}{}
		}

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
	fname := fmt.Sprintf("%s/%d_%s", "storage", n.GetID(), blockchainFile)

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
