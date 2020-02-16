package chain

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/golang/protobuf/ptypes"
)

//go:generate mockgen -destination=./mocks/hasher_mock.go -package=mocks github.com/asgaines/blockchain/chain Hasher
type Hasher interface {
	Hash(b *Block) []byte
}

type hasher struct{}

func NewHasher() Hasher {
	return &hasher{}
}

func (h *hasher) Hash(b *Block) []byte {
	nonceB := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceB, b.Nonce)

	headerB := []byte(ptypes.TimestampString(b.Timestamp))
	headerB = append(headerB, b.Prevhash...)
	headerB = append(headerB, nonceB...)
	headerB = append(headerB, b.Target...)
	headerB = append(headerB, b.MerkleRoot...)

	hb := sha256.Sum256(headerB)

	return hb[:]
}
