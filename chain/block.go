package chain

import (
	"crypto/sha256"
	"encoding/binary"
	"strconv"

	"github.com/golang/protobuf/ptypes"

	pb "github.com/asgaines/blockchain/protogo/blockchain"
)

// Block is a wrapper for the protobuf block representation: a link in the blockchain
type Block pb.Block

// NewBlock instantiates a Block from a payload
func NewBlock(prev *Block, txs []*pb.Tx, nonce uint64, target uint64, pubkey string) *Block {
	b := Block{
		Timestamp: ptypes.TimestampNow(),
		Prevhash:  prev.Hash,
		Nonce:     nonce,
		Target:    target,
		Pubkey:    pubkey,
		Txs:       txs,
	}

	b.Hash = b.makeHash()

	return &b
}

func (b *Block) ToProto() *pb.Block {
	return (*pb.Block)(b)
}

func (b Block) makeHash() uint64 {
	concat := ptypes.TimestampString(b.Timestamp)
	concat += strconv.FormatUint(b.Prevhash, 10)
	concat += strconv.FormatUint(b.Nonce, 10)
	concat += strconv.FormatUint(b.Target, 10)
	concat += b.Pubkey

	for _, tx := range b.Txs {
		concat += strconv.FormatUint(tx.Hash, 10)
	}

	h := sha256.New()
	h.Write([]byte(concat))
	hash := h.Sum(nil)

	// Only take most significant 8 bytes
	// A full implementation uses a 256 bit number,
	// but limiting here to Go builtin type for ease
	top8 := hash[:8]

	return binary.BigEndian.Uint64(top8)
}
