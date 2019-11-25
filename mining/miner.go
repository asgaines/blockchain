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
// const MaxTarget float64 = 0xFF_FF_FF_FF_FF_FF_FF_FF

//go:generate mockgen -destination=./mocks/miner_mock.go -package=mocks github.com/asgaines/blockchain/mining Miner
type Miner interface {
	Mine(ctx context.Context, mineshaft chan<- *chain.Block)
	AddTx(tx *pb.Tx)
	SetPrevBlock(block *chain.Block)
	SetTarget(target float64)
	ClearTxs()
}

func NewMiner(prevBlock *chain.Block, pubkey string, difficulty float64, targetDurPerBlock time.Duration, hashSpeed HashSpeed, target float64, hasher chain.Hasher) Miner {
	m := miner{
		prevBlock: prevBlock,
		pubkey:    pubkey,
		// difficulty:        difficulty,
		// targetDurPerBlock: targetDurPerBlock,
		hashSpeed: hashSpeed,
		target:    target,
		hasher:    hasher,
	}

	// m.target = m.calcTarget(difficulty)

	return &m
}

type miner struct {
	prevBlock *chain.Block
	pubkey    string
	target    float64
	// targetDurPerBlock time.Duration
	nonce     uint64
	hashSpeed HashSpeed
	txpool    []*pb.Tx
	hasher    chain.Hasher
	// difficulty        float64
}

func (m *miner) Mine(ctx context.Context, conveyor chan<- *chain.Block) {
	log.Printf("%f (target)", m.target)
	defer close(conveyor)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// switch m.hashSpeed {
		// case LowSpeed:
		// 	time.Sleep(100 * time.Millisecond)
		// case MediumSpeed:
		// 	time.Sleep(10 * time.Millisecond)
		// case HighSpeed:
		// 	time.Sleep(1 * time.Millisecond)
		// case UltraSpeed:
		// }

		candidate := chain.NewBlock(
			m.hasher,
			m.prevBlock,
			m.txpool,
			m.nonce,
			m.target,
			m.pubkey,
		)

		if solved := float64(candidate.Hash) <= m.target; solved {
			conveyor <- candidate
			m.prevBlock = candidate
			m.nonce = 0
		} else {
			m.nonce++
		}
	}
}

func (m *miner) SetPrevBlock(block *chain.Block) {
	m.prevBlock = block
}

func (m *miner) SetTarget(target float64) {
	m.target = target
}

func (m *miner) AddTx(tx *pb.Tx) {
	m.txpool = append(m.txpool, tx)
}

func (m *miner) ClearTxs() {
	// TODO: ensure ALL txs in txpool are in new chain
	// If not, keep orphans in txpool
	m.txpool = []*pb.Tx{}
}
