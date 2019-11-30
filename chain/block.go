package chain

import (
	"github.com/golang/protobuf/ptypes"

	pb "github.com/asgaines/blockchain/protogo/blockchain"
)

// Block is a wrapper for the protobuf block representation: a link in the blockchain
type Block pb.Block

// NewBlock instantiates a Block from a payload
func NewBlock(hasher Hasher, prev *Block, txs []*pb.Tx, nonce uint64, target []byte, pubkey string) *Block {
	b := &Block{
		Timestamp: ptypes.TimestampNow(),
		Prevhash:  prev.Hash,
		Nonce:     nonce,
		Target:    target,
		Pubkey:    pubkey,
		Txs:       txs,
	}

	b.Hash = hasher.Hash(b)

	return b
}

func (b *Block) ToProto() *pb.Block {
	return (*pb.Block)(b)
}
