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
	payload += b.Pubkey

	for _, tx := range b.Txs {
		payload += strconv.FormatUint(tx.Hash, 10)
	}

	hh := sha256.New()
	hh.Write([]byte(payload))
	hb := hh.Sum(nil)

	// Only take most significant 8 bytes
	// A full implementation uses a 256 bit number,
	// but limiting here to Go builtin type for ease
	// top8 := hb[:8]

	// return new(big.Int).SetBytes(hb) //binary.BigEndian.Uint64(top8)
	return hb
}
