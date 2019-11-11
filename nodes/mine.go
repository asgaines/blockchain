package nodes

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/asgaines/blockchain/chain"
	"github.com/golang/protobuf/ptypes"
)

func (n *node) mine(ctx context.Context) {
	mineshaft := make(chan *chain.Block)

	go n.miner.Mine(ctx, mineshaft)

	for {
		select {
		case minedBlock := <-mineshaft:
			chain := n.chain.AddBlock(minedBlock)
			if overridden := n.setChain(chain); !overridden {
				log.Fatal("solving a block did not successfully lead to chain override")
			}
		case <-ctx.Done():
			return
		}
	}
}

func (n *node) logBlock(block *chain.Block) {
	// lastLinkDur, err := n.getLastBlockDur(n.chain)
	// if err != nil {
	// 	log.Println(err)
	// }

	// log.Printf("%064b (%vs)\n", block.Hash, lastLinkDur.Seconds())

	ave := time.Since(n.startTime).Seconds() / float64(n.chain.Length()-1)

	if _, err := n.ratesF.Write([]byte(fmt.Sprintf("%v\n", ave))); err != nil {
		log.Printf("could not write to file: %s", err)
	}
}

func (n *node) setChain(chain *chain.Chain) bool {
	if chain.IsSolid() && chain.Length() > n.chain.Length() {
		n.chain = chain

		n.logBlock(chain.LastLink())

		if (chain.Length()-1)%n.recalcPeriod == 0 {
			actualAvgBlockDur, err := n.getRangeAvgBlockDur(n.chain, n.recalcPeriod)
			if err != nil {
				log.Println(err)
			}

			n.miner.RecalcTarget(actualAvgBlockDur)
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
		return 0, fmt.Errorf("not enough solves for recalc period. chain length: %d, recalc period: %d", c.Length(), recalcPeriod)
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
