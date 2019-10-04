package cluster

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/asgaines/blockchain/chain"
	"github.com/asgaines/blockchain/nodes"
	"github.com/asgaines/blockchain/transactions"
)

// InitialExpectedHashrate is a guess of how many hashes are possible per second.
// It seeds the network for the first block, but replaced by real data once it
// comes through.
//
// Do not set it unreasonably high, or the first block could never be solved.
const InitialExpectedHashrate = float64(1_6) //00_000)

// Cluster represents a set of Nodes
// It's useful for coordinating behavior across blockchain network
type Cluster interface {
	Run(ctx context.Context)
	AddTransaction(tx transactions.Tx)
	UpdateTarget()
	UpdateHashrate()
	ListenForSolves()
	WriteRates()
}

// NewCluster instantiates a Cluster; a set of Nodes
func NewCluster(numNodes int, targetDur time.Duration, recalcPeriod int) Cluster {
	chainSeed := chain.InitBlockchain()

	c := cluster{
		network:      make([]nodes.Node, 0, numNodes),
		wg:           &sync.WaitGroup{},
		hashReport:   make(chan uint64),
		targetDur:    targetDur,
		hashesPerSec: InitialExpectedHashrate,
		lastSolve:    time.Now(),
		recalcPeriod: recalcPeriod,
	}

	for id := 0; id < numNodes; id++ {
		solve := make(chan nodes.SolveReport)
		node := nodes.NewNode(
			id,
			chainSeed,
			solve,
			c.calcTarget(),
		)
		c.solves = append(c.solves, solve)
		c.network = append(c.network, node)
	}

	for i, node := range c.network {
		peers := []nodes.Node{c.network[(i+1)%len(c.network)]}
		node.SetPeers(peers)
	}

	return &c
}

type cluster struct {
	network      []nodes.Node
	wg           *sync.WaitGroup
	hashReport   chan uint64
	numHashes    uint64
	startTime    time.Time
	targetDur    time.Duration
	hashesPerSec float64
	solves       []<-chan nodes.SolveReport
	lastSolve    time.Time
	totalSolves  int
	rates        []float64
	recalcPeriod int
}

func (c *cluster) Run(ctx context.Context) {
	defer c.WriteRates()

	c.startTime = time.Now()

	c.wg.Add(1)
	go func() {
		c.ListenForSolves()
		c.wg.Done()
	}()

	c.wg.Add(len(c.network))
	for _, node := range c.network {
		go func(node nodes.Node) {
			defer c.wg.Done()
			node.Run(ctx)
		}(node)
	}

	c.wg.Wait()
}

func (c *cluster) ListenForSolves() {
	for solve := range c.mergeReports(c.solves...) {
		c.totalSolves++

		fmt.Printf("Node ID %d solved new block!\n", solve.ID)
		fmt.Printf("%064b\n", solve.Block.Target)
		fmt.Printf("%064b\n", solve.Block.Hash)

		fmt.Printf("Took %v\n", time.Since(c.lastSolve))
		c.lastSolve = time.Now()

		ave := time.Since(c.startTime).Seconds() / float64(c.totalSolves)
		fmt.Printf("Average %v seconds/block\n", ave)
		c.rates = append(c.rates, ave)

		fmt.Println()

		if c.totalSolves%c.recalcPeriod == 0 {
			c.UpdateTarget()
		}
	}
}

func (c cluster) AddTransaction(tx transactions.Tx) {
	for _, node := range c.network {
		node.SubmitTransaction(tx)
	}
}

// UpdateTarget calculates a new mining target (block hash must be less than target
// to be considered solved), and distributes it to the network
func (c *cluster) UpdateTarget() {
	c.UpdateHashrate()
	target := c.calcTarget()

	for _, n := range c.network {
		n.UpdateTarget(target)
	}
}

func (c *cluster) UpdateHashrate() {
	numHashes := uint64(0)

	for _, node := range c.network {
		numHashes += node.GetNumHashes()
	}

	// This is overall average. Likely wanting a more instantaneous rate to
	// respond to changes more rapidly
	c.hashesPerSec = float64(numHashes) / time.Since(c.startTime).Seconds()
}

func (c *cluster) WriteRates() {
	buf := new(bytes.Buffer)
	if _, err := buf.Write([]byte("sec/block\n")); err != nil {
		log.Println(err)
	}
	for _, rate := range c.rates {
		if _, err := buf.Write([]byte(fmt.Sprintf("%v\n", rate))); err != nil {
			log.Println(err)
		}
	}
	if err := ioutil.WriteFile("rates.out", buf.Bytes(), 0644); err != nil {
		log.Println(err)
	}
}

// calcTarget retrieves a target which must be greater than a hash value
// for a block to be considered solved.
// It is a determinant for how much work is expected for a
// new block to be added to chain.
func (c *cluster) calcTarget() uint64 {
	// Lowest possible difficulty (highest possible target)
	var difficulty1Target uint64 = 0xFF_FF_FF_FF_FF_FF_FF_FF

	shifts := math.Log2(c.hashesPerSec * c.targetDur.Seconds())
	target := difficulty1Target >> int(math.Round(shifts))

	return target
}

func (c *cluster) mergeReports(chans ...<-chan nodes.SolveReport) <-chan nodes.SolveReport {
	master := make(chan nodes.SolveReport)

	go func() {
		var wg sync.WaitGroup

		wg.Add(len(chans))
		for _, c := range chans {
			go func(c <-chan nodes.SolveReport) {
				for solve := range c {
					master <- solve
				}
				wg.Done()
			}(c)
		}

		wg.Wait()
		close(master)
	}()

	return master
}

func (c cluster) randomNode() nodes.Node {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	return c.network[r.Intn(len(c.network))]
}
