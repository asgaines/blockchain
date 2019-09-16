package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

// Node represents a blockchain client
// It preserves a copy of the blockchain, competes for new block additions,
// and verifies work of peer nodes
type Node interface {
	GetID() int
	Run(context.Context)
	SubmitTransaction(lb LilBits, podium chan<- Node)
	Mine(lb LilBits)
	GetCC() chan Blockchain
	// AddLilBits(LilBits) error
	GetChain() *Blockchain
	SetChain(Blockchain) error
	SetPeers([]Node)
	SenseDifficulty() float64
}

// NewNode instantiates a Node; a blockchain client
func NewNode(id int, chain Blockchain, transactions chan LilBits) Node {
	return &node{
		id:           id,
		chain:        chain,
		transactions: transactions,
		cc:           make(chan Blockchain),
		latency:      100 * time.Millisecond,
		ilost:        make(chan struct{}),
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
}

func (n *node) Run(ctx context.Context) {
	for {
		select {
		case lilbits := <-n.transactions:
			fmt.Printf("%d received new submission request: %v\n", n.id, lilbits)
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
	nonce := rand.Intn(1000)
	orig := nonce

	difficulty := n.SenseDifficulty()

	for {
		block := NewBlock(n.chain[len(n.chain)-1], lb, nonce)

		match := true
		for i := 0; i < 5+int(difficulty*5); i++ {
			if block.Hash[i] != '0' {
				match = false
				break
			}
		}

		if match {
			fmt.Printf("Node %d mined new block of hash %s. %d nonce updates.\n", n.id, block.Hash, nonce-orig)
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
	if !chain.IsBroken() && len(chain) > len(n.chain) {
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

// SenseDifficulty retieves a value between 0 and 1 where 1 is maximum difficulty
// difficulty is used to determine how much work to be expected for a new block to be added
// to chain.
func (n *node) SenseDifficulty() float64 {
	rand.Seed(int64(n.id) * time.Now().Unix())
	return rand.Float64()
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
