package transactions

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"github.com/golang/protobuf/ptypes"
)

func SetHash(tx *pb.Tx) {
	concat := fmt.Sprintf("%f", tx.Value)
	concat += ptypes.TimestampString(tx.Timestamp)
	concat += tx.For
	concat += tx.From
	concat += tx.To

	h := sha256.New()
	h.Write([]byte(concat))
	hash := h.Sum(nil)

	top8 := hash[0:8]

	tx.Hash = binary.BigEndian.Uint64(top8)
}
