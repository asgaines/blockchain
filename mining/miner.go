package mining

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/asgaines/blockchain/chain"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
)

// MaxTarget is the highest possible target value (lowest possible difficulty)
// It is the highest potential hash output of sha256 (2**256 - 1), used for calculating the
// positioning of the target according to current difficulty.
// As difficulty increases, target decreases.
var MaxTarget = new(big.Int).SetBytes([]byte{
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255,
})

//go:generate mockgen -destination=./mocks/miner_mock.go -package=mocks github.com/asgaines/blockchain/mining Miner
type Miner interface {
	Mine(ctx context.Context, mineshaft chan<- BlockReport)
	AddTx(tx *pb.Tx)
	SetPrevBlock(block *chain.Block)
	SetTarget(difficulty float64) error
	ClearTxs()
}

type BlockReport struct {
	ID    int
	Block *chain.Block
}

// NewMiner returns an implementation of Miner, ready to begin mining
func NewMiner(ID int, prevBlock *chain.Block, pubkey string, difficulty float64, targetDurPerBlock time.Duration, hashSpeed HashSpeed, hasher chain.Hasher) Miner {
	m := miner{
		ID:        ID,
		prevBlock: prevBlock,
		pubkey:    pubkey,
		hashSpeed: hashSpeed,
		hasher:    hasher,
	}

	if err := m.SetTarget(difficulty); err != nil {
		log.Fatal(err)
	}

	return &m
}

type miner struct {
	ID        int
	prevBlock *chain.Block
	pubkey    string
	target    []byte
	nonce     uint64
	hashSpeed HashSpeed
	txpool    []*pb.Tx
	hasher    chain.Hasher
}

func (m *miner) Mine(ctx context.Context, conveyor chan<- BlockReport) {
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

		hb := new(big.Int).SetBytes(candidate.Hash)
		tb := new(big.Int).SetBytes(m.target)

		cmp := hb.Cmp(tb)
		if solved := cmp == 0 || cmp == -1; solved {
			conveyor <- BlockReport{
				ID:    m.ID,
				Block: candidate,
			}
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

func (m *miner) SetTarget(difficulty float64) error {
	if difficulty < 1 {
		return fmt.Errorf("minimum difficulty is 1, cannot set target based on value %v", difficulty)
	}

	diffF := new(big.Float).SetFloat64(difficulty)

	target, _ := new(big.Float).Quo(new(big.Float).SetInt(MaxTarget), diffF).Int(nil)

	if target.Cmp(MaxTarget) == 1 {
		m.target = new(big.Int).Set(MaxTarget).Bytes()
		return nil
	}

	m.target = target.Bytes()
	log.Printf("%064x (target)", m.target)

	return nil
}

func (m *miner) AddTx(tx *pb.Tx) {
	m.txpool = append(m.txpool, tx)
}

func (m *miner) ClearTxs() {
	// TODO: ensure ALL txs in txpool are in new chain
	// If not, keep orphans in txpool
	m.txpool = []*pb.Tx{}
}
