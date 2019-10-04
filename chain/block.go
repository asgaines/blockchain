package chain

import (
	"crypto/sha256"
	"encoding/binary"
	"strconv"
	"time"

	pb "github.com/asgaines/blockchain/protogo/blockchain"
)

// Block represents a link in the blockchain
type Block struct {
	Timestamp int64    `json:"timestamp"`
	Hash      uint64   `json:"hash"`
	PrevHash  uint64   `json:"prevhash"`
	Nonce     uint64   `json:"nonce"`
	Target    uint64   `json:"target"`
	Recipient string   `json:"recipient"`
	Txs       []*pb.Tx `json:"txs"`
}

// NewBlock instantiates a Block from a payload
func NewBlock(prev *Block, txs []*pb.Tx, nonce uint64, target uint64, recipient string) *Block {
	b := Block{
		Timestamp: time.Now().UTC().UnixNano(),
		PrevHash:  prev.Hash,
		Nonce:     nonce,
		Target:    target,
		Recipient: recipient,
		Txs:       txs,
	}

	b.Hash = b.makeHash()

	return &b
}

func (b Block) makeHash() uint64 {
	concat := strconv.FormatInt(b.Timestamp, 10)
	concat += strconv.FormatUint(b.PrevHash, 10)
	concat += strconv.FormatUint(b.Nonce, 10)
	concat += strconv.FormatUint(b.Target, 10)
	concat += b.Recipient

	for _, tx := range b.Txs {
		concat += tx.String()
	}

	h := sha256.New()
	h.Write([]byte(concat))
	hashed := h.Sum(nil)

	// Only take most significant 8 bytes
	// A full implementation uses a 256 bit number,
	// but limiting here to Go builtin type for ease
	top8 := hashed[0:8]

	return binary.BigEndian.Uint64(top8)
}
