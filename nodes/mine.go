package nodes

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/asgaines/blockchain/chain"
	"github.com/asgaines/blockchain/mining"
	"github.com/golang/protobuf/ptypes"
)

// DiffConfineFactor is the maximum amount by which a difficulty value can be
// adjusted by any one recalculation
const DiffConfineFactor float64 = 4

func (n *node) mine(ctx context.Context) {
	conveyors := make([]<-chan mining.BlockReport, 0, len(n.miners))

	for _, miner := range n.miners {
		conveyor := make(chan mining.BlockReport)
		conveyors = append(conveyors, conveyor)

		go miner.Mine(ctx, conveyor)
	}

	for mineReport := range n.mergeConveyors(conveyors...) {
		chain := n.chain.WithBlock(mineReport.Block)
		if overridden := n.setChain(chain, true); !overridden {
			log.Fatal("solving a block did not successfully lead to own chain override")
		}
		n.propagateChain(nil)
	}
}

func (n *node) mergeConveyors(conveyors ...<-chan mining.BlockReport) <-chan mining.BlockReport {
	mergedConveyors := make(chan mining.BlockReport)

	go func() {
		var wg sync.WaitGroup

		wg.Add(len(conveyors))
		for _, c := range conveyors {
			go func(c <-chan mining.BlockReport) {
				for solve := range c {
					mergedConveyors <- solve
				}
				wg.Done()
			}(c)
		}

		wg.Wait()
		close(mergedConveyors)
	}()

	return mergedConveyors
}

func (n *node) logBlock(block *chain.Block) {
	lastLinkDur, err := n.getLastBlockDur(n.chain)
	if err != nil {
		log.Println(err)
	}

	if _, err := n.dursF.Write([]byte(fmt.Sprintf("%v\t%v\n", lastLinkDur.Seconds(), n.difficulty))); err != nil {
		log.Printf("could not write to file: %s", err)
	}

	minedBy := block.GetMinerPubkey()
	if minedBy == n.pubkey {
		minedBy = fmt.Sprintf("%s (you)", minedBy)
	}

	log.Printf("%064x (%vs) [%s]\n", block.Hash, lastLinkDur.Seconds(), minedBy)
}

func (n *node) setChain(chain *chain.Chain, trusted bool) bool {
	if (trusted || n.IsValid(chain)) && chain.Length() > n.chain.Length() {
		n.chain = chain
		n.updatePrevBlock(chain.LastLink())

		n.logBlock(chain.LastLink())

		if (chain.Length()-1)%n.recalcPeriod == 0 {
			actualAvgBlockDur, err := n.getRangeAvgBlockDur(n.chain, n.recalcPeriod)
			if err != nil {
				log.Println(err)
			}

			if _, err := n.statsF.Write([]byte(fmt.Sprintf("%v\t%v\n", actualAvgBlockDur.Seconds(), n.difficulty))); err != nil {
				log.Printf("could not write to file: %s", err)
			}

			n.difficulty = n.calcDifficulty(actualAvgBlockDur, n.difficulty)

			n.updateTarget(n.difficulty)
		}

		n.resetTxpool()
		return true
	}

	return false
}

func (n *node) updateTarget(difficulty float64) {
	for _, miner := range n.miners {
		miner.SetTarget(difficulty)
	}
}

func (n *node) updatePrevBlock(block *chain.Block) {
	for _, miner := range n.miners {
		miner.SetPrevBlock(block)
	}
}

func (n *node) IsValid(c *chain.Chain) bool {
	if len(c.Pbc.Blocks) <= 0 {
		return false
	}

	for i, block := range c.Pbc.Blocks[1:] {
		prev := c.Pbc.Blocks[i]

		prevhash := n.hasher.Hash((*chain.Block)(prev))
		if !bytes.Equal(prevhash, block.Prevhash) ||
			!bytes.Equal(prevhash, prev.Hash) ||
			!bytes.Equal(block.Hash, n.hasher.Hash((*chain.Block)(block))) ||
			bytes.Compare(block.Hash, block.Target) == 1 {
			return false
		}
	}

	return true
}

func (n *node) getRecalcRangeDur(c *chain.Chain, recalcPeriod int) (time.Duration, error) {
	if recalcPeriod > c.Length()-1 {
		return 0, fmt.Errorf("not enough blocks for recalc period. chain length: %d, recalc period: %d", c.Length(), recalcPeriod)
	}

	latestBlockTime, err := ptypes.Timestamp(c.LastLink().Timestamp)
	if err != nil {
		return 0, nil
	}

	lastRecalcTime := c.BlockByIdx(c.Length() - 1 - recalcPeriod).Timestamp

	prevRecalc, err := ptypes.Timestamp(lastRecalcTime)
	if err != nil {
		return 0, err
	}

	return latestBlockTime.Sub(prevRecalc), nil
}

func (n *node) getLastBlockDur(c *chain.Chain) (time.Duration, error) {
	return n.getRecalcRangeDur(c, 1)
}

func (n *node) getRangeAvgBlockDur(c *chain.Chain, recalcPeriod int) (time.Duration, error) {
	actualRangeDur, err := n.getRecalcRangeDur(c, recalcPeriod)
	if err != nil {
		return 0, err
	}

	return actualRangeDur / time.Duration(recalcPeriod), nil
}

func (n *node) calcDifficulty(actualDurPerBlock time.Duration, currDifficulty float64) float64 {
	adjustment := float64(n.targetDurPerBlock) / float64(actualDurPerBlock)
	adjustment = n.confine(adjustment)

	newDifficulty := currDifficulty * adjustment

	return newDifficulty
}

// confine restricts the difficulty adjustment shift to a factor of +/-4 to protect
// against anomalous fluctuations
func (n *node) confine(adjustment float64) float64 {
	confined := math.Max(1/DiffConfineFactor, adjustment)
	confined = math.Min(DiffConfineFactor, confined)

	return confined
}
