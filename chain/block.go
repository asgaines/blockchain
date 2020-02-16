package chain

import (
	"crypto/sha256"

	"github.com/golang/protobuf/ptypes"

	pb "github.com/asgaines/blockchain/protogo/blockchain"
)

// Block is a wrapper for the protobuf block representation: a link in the blockchain
type Block pb.Block

// NewBlock instantiates a Block from a payload
func NewBlock(hasher Hasher, prev *Block, txs []*pb.Tx, nonce uint64, target []byte, pubkey string) *Block {
	txHashes := []byte{}
	for _, tx := range txs {
		txHashes = append(txHashes, tx.Hash...)
	}

	// A simplified merkle root of all transactions
	merkleRoot := sha256.Sum256(txHashes)

	b := &Block{
		Timestamp:  ptypes.TimestampNow(),
		Prevhash:   prev.Hash,
		Nonce:      nonce,
		Target:     target,
		MerkleRoot: merkleRoot[:],
		Txs:        txs,
	}

	b.Hash = hasher.Hash(b)

	return b
}

func (b *Block) GetMinerPubkey() string {
	var pubkey string

	for _, tx := range b.Txs {
		if tx.GetSender() == "" {
			pubkey = tx.GetRecipient()
			break
		}
	}

	return pubkey
}

func (b *Block) ToProto() *pb.Block {
	return (*pb.Block)(b)
}
