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
	UpdatePrevHash(hash []byte)
	SetTarget(difficulty float64) error
	SetTxs(txs []*pb.Tx)
}

type BlockReport struct {
	ID    int
	Block *chain.Block
}

// NewMiner returns an implementation of Miner. It still requires setting the target
// from the difficulty and the previous block before mining.
func NewMiner(ID int, pubkey string, targetDurPerBlock time.Duration, hashSpeed HashSpeed, hasher chain.Hasher) Miner {
	m := miner{
		ID:        ID,
		pubkey:    pubkey,
		hashSpeed: hashSpeed,
		hasher:    hasher,
	}

	return &m
}

type miner struct {
	ID        int
	prevHash  []byte
	pubkey    string
	target    []byte
	nonce     uint64
	hashSpeed HashSpeed
	txs       []*pb.Tx
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

		candidate := chain.NewBlock(
			m.hasher,
			m.prevHash,
			m.txs,
			m.nonce,
			m.target,
			m.pubkey,
		)

		hash := m.hasher.Hash(candidate)

		hashBI := new(big.Int).SetBytes(hash)
		targetBI := new(big.Int).SetBytes(m.target)

		// Block is considered solved if the generated hash is less than or equal
		// to the target value
		cmp := hashBI.Cmp(targetBI)
		if solved := cmp != 1; solved {
			conveyor <- BlockReport{
				ID:    m.ID,
				Block: candidate,
			}
			m.UpdatePrevHash(hash[:])
		} else {
			m.nonce++
		}
	}
}

func (m *miner) UpdatePrevHash(hash []byte) {
	m.prevHash = hash
	m.nonce = 0
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

func (m *miner) SetTxs(txs []*pb.Tx) {
	m.txs = txs
}
