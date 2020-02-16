package chain

import (
	"crypto/sha256"
	"strconv"

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
	payload := ptypes.TimestampString(b.Timestamp)
	payload += string(b.Prevhash)
	payload += strconv.FormatUint(b.Nonce, 10)
	payload += string(b.Target)
	payload += string(b.MerkleRoot)

	hh := sha256.New()
	hh.Write([]byte(payload))
	hb := hh.Sum(nil)

	// return new(big.Int).SetBytes(hb) //binary.BigEndian.Uint64(top8)
	return hb
}
