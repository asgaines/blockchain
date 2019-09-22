package main

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// Cluster represents a set of Nodes
// It's useful for coordinating behavior across blockchain network
type Cluster interface {
	Run(context.Context)
	AddTransaction(LilBits) (Node, error)
}

// NewCluster instantiates a Cluster; a set of Nodes
func NewCluster(numNodes int, chainSeed Blockchain, targetMin float64) Cluster {
	var nodes []Node

	for id := 0; id < numNodes; id++ {
		submissions := make(chan LilBits)
		node := NewNode(id, chainSeed, submissions, targetMin)
		nodes = append(nodes, node)
	}

	for i, node := range nodes {
		peers := []Node{nodes[(i+1)%len(nodes)]}
		node.SetPeers(peers)
	}

	return &cluster{
		nodes: nodes,
		wg:    &sync.WaitGroup{},
	}
}

type cluster struct {
	nodes []Node
	wg    *sync.WaitGroup
}

func (c cluster) Run(ctx context.Context) {
	for _, node := range c.nodes {
		c.wg.Add(1)
		go func(node Node) {
			defer c.wg.Done()
			node.Run(ctx)
		}(node)
	}

	c.wg.Wait()
}

func (c cluster) AddTransaction(lb LilBits) (Node, error) {
	podium := make(chan Node)

	for _, node := range c.nodes {
		node.SubmitTransaction(lb, podium)
	}

	node := <-podium

	// Closing the podium opens the blockchain up for another mining node causing a panic
	// when writing its solution if it hasn't yet realized it lost
	// The panic will need to be solved by a way of reconciling near-simultaneous successful minings
	close(podium)

	return node, nil
}

func (c cluster) randomNode() Node {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	return c.nodes[r.Intn(len(c.nodes))]
}
