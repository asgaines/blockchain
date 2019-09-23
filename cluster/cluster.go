package cluster

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/asgaines/blockchain/blockchain"
	"github.com/asgaines/blockchain/nodes"
	"github.com/asgaines/blockchain/transactions"
)

// Cluster represents a set of Nodes
// It's useful for coordinating behavior across blockchain network
type Cluster interface {
	Run(ctx context.Context)
	AddTransaction(tx transactions.Transaction)
}

// NewCluster instantiates a Cluster; a set of Nodes
func NewCluster(numNodes int, chainSeed blockchain.Blockchain, targetMin float64) Cluster {
	var network []nodes.Node

	for id := 0; id < numNodes; id++ {
		submissions := make(chan transactions.Transaction)
		node := nodes.NewNode(id, chainSeed, submissions, targetMin)
		network = append(network, node)
	}

	for i, node := range network {
		peers := []nodes.Node{network[(i+1)%len(network)]}
		node.SetPeers(peers)
	}

	return &cluster{
		network: network,
		wg:      &sync.WaitGroup{},
	}
}

type cluster struct {
	network []nodes.Node
	wg      *sync.WaitGroup
}

func (c cluster) Run(ctx context.Context) {
	for _, node := range c.network {
		c.wg.Add(1)
		go func(node nodes.Node) {
			defer c.wg.Done()
			node.Run(ctx)
		}(node)
	}

	c.wg.Wait()
}

func (c cluster) AddTransaction(tx transactions.Transaction) {
	for _, node := range c.network {
		node.SubmitTransaction(tx)
	}
}

func (c cluster) randomNode() nodes.Node {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	return c.network[r.Intn(len(c.network))]
}
