package chain

import (
	"crypto/sha256"
	"encoding/binary"
	"strconv"
	"time"

	"github.com/asgaines/blockchain/transactions"
)

// Block represents a link in the blockchain
type Block struct {
	Timestamp    int64            `json:"timestamp"`
	Hash         uint64           `json:"hash"`
	PrevHash     uint64           `json:"prevhash"`
	Nonce        uint64           `json:"nonce"`
	Target       uint64           `json:"target"`
	Transactions transactions.Txs `json:"payload"`
}

// NewBlock instantiates a Block from a payload
func NewBlock(prev *Block, txs transactions.Txs, nonce uint64, target uint64) *Block {
	b := Block{
		Timestamp:    time.Now().UTC().UnixNano(),
		PrevHash:     prev.Hash,
		Nonce:        nonce,
		Target:       target,
		Transactions: txs,
	}

	b.Hash = b.makeHash()

	return &b
}

func (b Block) makeHash() uint64 {
	concat := strconv.FormatInt(b.Timestamp, 10)
	concat += strconv.FormatUint(b.PrevHash, 10)
	concat += b.Transactions.String()
	concat += strconv.FormatUint(b.Nonce, 10)
	concat += strconv.FormatUint(b.Target, 10)

	h := sha256.New()
	h.Write([]byte(concat))
	hashed := h.Sum(nil)

	// Only take most significant 8 bytes
	// A full implementation uses a 256 bit number,
	// but limiting here to Go builtin type for ease
	top8 := hashed[0:8]

	return binary.BigEndian.Uint64(top8)
}
