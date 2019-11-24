package chain

import (
	"crypto/sha256"
	"encoding/binary"
	"strconv"

	"github.com/golang/protobuf/ptypes"
)

type Hasher interface {
	Hash(b *Block) uint64
}

type hasher struct{}

//go:generate mockgen -destination=./mocks/hasher_mock.go -package=mocks github.com/asgaines/blockchain/chain Hasher
func NewHasher() Hasher {
	return &hasher{}
}

func (h *hasher) Hash(b *Block) uint64 {
	concat := ptypes.TimestampString(b.Timestamp)
	concat += strconv.FormatUint(b.Prevhash, 10)
	concat += strconv.FormatUint(b.Nonce, 10)
	concat += strconv.FormatFloat(b.Target, 'E', -1, 64)
	concat += b.Pubkey

	for _, tx := range b.Txs {
		concat += strconv.FormatUint(tx.Hash, 10)
	}

	hh := sha256.New()
	hh.Write([]byte(concat))
	hb := hh.Sum(nil)

	// Only take most significant 8 bytes
	// A full implementation uses a 256 bit number,
	// but limiting here to Go builtin type for ease
	top8 := hb[:8]

	return binary.BigEndian.Uint64(top8)
}
