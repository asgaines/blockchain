package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/asgaines/blockchain/transactions"
)

// Block represents a link in the blockchain
type Block struct {
	ID           int                       `json:"ID"`
	Timestamp    int64                     `json:"timestamp"`
	Hash         string                    `json:"hash"`
	PrevHash     string                    `json:"prevhash"`
	Nonce        int                       `json:"nonce"`
	Transactions transactions.Transactions `json:"payload"`
}

// NewBlock instantiates a Block from a payload
func NewBlock(prev *Block, txs transactions.Transactions, nonce int) *Block {
	b := Block{
		ID:           prev.ID + 1,
		Timestamp:    time.Now().UTC().UnixNano(),
		PrevHash:     prev.Hash,
		Nonce:        nonce,
		Transactions: txs,
	}

	b.Hash = b.makeHash()

	return &b
}

func (b Block) makeHash() string {
	concat := string(b.ID) + strconv.FormatInt(b.Timestamp, 10) + b.PrevHash + b.Transactions.String() + strconv.Itoa(b.Nonce)

	h := sha256.New()
	h.Write([]byte(concat))
	hashed := h.Sum(nil)

	return hex.EncodeToString(hashed)
}
