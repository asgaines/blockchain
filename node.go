package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Node interface {
	GetID() int
	Run(context.Context, *sync.WaitGroup)
	GetCC() chan Blockchain
	AddLilBits(LilBits) error
	GetChain() Blockchain
	SetChain(Blockchain) error
	SetPeers([]Node)
}

func NewNode(id int, chain Blockchain) Node {
	return &node{
		id:    id,
		chain: chain,
		cc:    make(chan Blockchain),
	}
}

type node struct {
	id    int
	peers []Node
	chain Blockchain
	cc    chan Blockchain
}

func (n node) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case chain := <-n.cc:
			fmt.Printf("%d: Received chain posting from peer %v\n", n.id, chain)
			if err := n.SetChain(chain); err != nil {
				log.Fatal(err)
			}
		case <-ctx.Done():
			log.Printf("%d: Shutting down\n", n.id)
			return
		}
	}
}

func (n node) GetCC() chan Blockchain {
	return n.cc
}

func (n node) GetID() int {
	return n.id
}

func (n *node) AddLilBits(lb LilBits) error {
	chain := n.chain.addBlock(lb)

	if err := n.SetChain(chain); err != nil {
		return err
	}

	return nil
}

func (n *node) GetChain() Blockchain {
	return n.chain
}

func (n *node) SetChain(chain Blockchain) error {
	if !chain.IsBroken() && len(chain) > len(n.chain) {
		fmt.Printf("%d: Overriding with new chain\n", n.id)
		n.chain = chain

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

func (n node) propagate() {
	fmt.Println(n.peers)
	for _, peer := range n.peers {
		fmt.Printf("%d: Propagation to %d\n\n", n.GetID(), peer.GetID())
		time.Sleep(2 * time.Second)

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
