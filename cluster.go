package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Cluster interface {
	Run(context.Context, *sync.WaitGroup)
	CompeteForWork(LilBits) (Blockchain, error)
}

func NewCluster(numNodes int, chainSeed Blockchain) Cluster {
	var nodes []Node

	for id := 0; id < numNodes; id++ {
		node := NewNode(id, chainSeed)
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

func (c cluster) Run(ctx context.Context, wg *sync.WaitGroup) {
	for _, node := range c.nodes {
		c.wg.Add(1)
		go node.Run(ctx, c.wg)
	}
	c.wg.Wait()
	wg.Done()
}

func (c cluster) CompeteForWork(lb LilBits) (Blockchain, error) {
	node := c.randomNode()
	if err := node.AddLilBits(lb); err != nil {
		return nil, err
	}

	fmt.Printf("NEW BLOCK SOLVED BY NODE %d\n", node.GetID())
	return node.GetChain(), nil
}

func (c cluster) randomNode() Node {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	return c.nodes[r.Intn(len(c.nodes))]
}
