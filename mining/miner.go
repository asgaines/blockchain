package mining

import (
	"context"
	"log"
	"time"

	"github.com/asgaines/blockchain/chain"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
)

// MaxTarget is the highest possible target value (lowest possible difficulty)
// As difficulty increases, target decreases.
const MaxTarget uint64 = 0xFF_FF_FF_FF_FF_FF_FF_FF

type Miner interface {
	Mine(ctx context.Context, mineshaft chan<- *chain.Block)
	AddTx(tx *pb.Tx)
	RecalcTarget(actualDur time.Duration)
	ClearTxs()
}

func NewMiner(prevBlock *chain.Block, pubkey string, difficulty float64, targetDurPerBlock time.Duration, hashSpeed HashSpeed) Miner {
	m := miner{
		prevBlock:         prevBlock,
		pubkey:            pubkey,
		difficulty:        difficulty,
		targetDurPerBlock: targetDurPerBlock,
		hashSpeed:         hashSpeed,
	}

	m.target = m.calcTarget(difficulty)

	return &m
}

type miner struct {
	prevBlock         *chain.Block
	pubkey            string
	target            uint64
	targetDurPerBlock time.Duration
	nonce             uint64
	hashSpeed         HashSpeed
	txpool            []*pb.Tx
	difficulty        float64
}

func (m *miner) Mine(ctx context.Context, mineshaft chan<- *chain.Block) {
	log.Printf("%064b (target)\n", m.target)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		switch m.hashSpeed {
		case LowSpeed:
			time.Sleep(100 * time.Millisecond)
		case MediumSpeed:
			time.Sleep(10 * time.Millisecond)
		case HighSpeed:
			time.Sleep(1 * time.Millisecond)
		case UltraSpeed:
		}

		blockCandidate := chain.NewBlock(
			m.prevBlock,
			m.txpool,
			m.nonce,
			m.target,
			m.pubkey,
		)

		m.nonce++

		if solved := blockCandidate.Hash <= m.target; solved {
			mineshaft <- blockCandidate
			m.prevBlock = blockCandidate
			m.nonce = 0
		}
	}
}

func (m *miner) AddTx(tx *pb.Tx) {
	m.txpool = append(m.txpool, tx)
}

func (m *miner) ClearTxs() {
	// TODO: ensure ALL txs in txpool are in new chain
	// If not, keep orphans in txpool
	m.txpool = []*pb.Tx{}
}

func (m *miner) RecalcTarget(actualAvgDur time.Duration) {
	newDifficulty := m.calcDifficulty(actualAvgDur, m.difficulty)
	// log.Printf("%v (new difficulty)", newDifficulty)
	m.difficulty = newDifficulty

	m.target = m.calcTarget(newDifficulty)
	// log.Println()
	// log.Printf("%064b (new target)\n", m.target)
}

func (m *miner) calcDifficulty(actualDurPerBlock time.Duration, currDifficulty float64) float64 {
	// log.Printf("actual dur per block: %v", actualDurPerBlock)
	adjustment := float64(m.targetDurPerBlock) / float64(actualDurPerBlock)
	// log.Printf("%v (adjustment)", adjustment)
	return currDifficulty * adjustment
}

func (m *miner) calcTarget(difficulty float64) uint64 {
	return uint64(float64(MaxTarget) / difficulty)
}
