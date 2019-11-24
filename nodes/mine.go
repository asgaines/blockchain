package nodes

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/asgaines/blockchain/chain"
	"github.com/golang/protobuf/ptypes"
)

// MaxTarget is the highest possible target value (lowest possible difficulty)
// As difficulty increases, target decreases.
const MaxTarget float64 = 0xFF_FF_FF_FF_FF_FF_FF_FF

func (n *node) mine(ctx context.Context) {
	conveyor := make(chan *chain.Block)

	go n.miner.Mine(ctx, conveyor)

	for {
		select {
		case minedBlock := <-conveyor:
			log.Printf("%#v", minedBlock)
			chain := n.chain.WithBlock(minedBlock)
			if overridden := n.setChain(chain, true); !overridden {
				log.Fatal("solving a block did not successfully lead to chain override")
			}
		case <-ctx.Done():
			return
		}
	}
}

func (n *node) logBlock(block *chain.Block) {
	lastLinkDur, err := n.getLastBlockDur(n.chain)
	if err != nil {
		log.Println(err)
	}

	if _, err := n.dursF.Write([]byte(fmt.Sprintf("%v\n", lastLinkDur.Seconds()))); err != nil {
		log.Printf("could not write to file: %s", err)
	}

	// log.Printf("%064b (%vs)\n", block.Hash, lastLinkDur.Seconds())
}

func (n *node) setChain(chain *chain.Chain, trusted bool) bool {
	if (trusted || chain.IsSolid(n.hasher)) && chain.Length() > n.chain.Length() {
		n.chain = chain

		n.logBlock(chain.LastLink())

		if (chain.Length()-1)%n.recalcPeriod == 0 {
			actualAvgBlockDur, err := n.getRangeAvgBlockDur(n.chain, n.recalcPeriod)
			if err != nil {
				log.Println(err)
			}

			target := n.recalcTarget(actualAvgBlockDur)
			n.miner.SetTarget(target)
		}

		n.miner.ClearTxs()
		// n.propagateChain()

		return true
	}

	log.Println("Not overriding")
	return false
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

func (n *node) recalcTarget(actualAvgDur time.Duration) float64 {
	newDifficulty := n.calcDifficulty(actualAvgDur, n.difficulty)
	// log.Printf("%v (new difficulty)", newDifficulty)
	n.difficulty = newDifficulty

	return math.Min(CalcTarget(newDifficulty), MaxTarget)
	// log.Println()
	// log.Printf("%064b (new target)\n", m.target)
}

func (n *node) calcDifficulty(actualDurPerBlock time.Duration, currDifficulty float64) float64 {
	// log.Printf("actual dur per block: %v", actualDurPerBlock)
	adjustment := float64(n.targetDurPerBlock) / float64(actualDurPerBlock)
	// log.Printf("%v (adjustment)", adjustment)
	return currDifficulty * adjustment
}

func CalcTarget(difficulty float64) float64 {
	return float64(MaxTarget) / difficulty
}
