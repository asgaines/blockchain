package transactions

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"github.com/golang/protobuf/ptypes"
)

func SetHash(tx *pb.Tx) {
	payload := fmt.Sprintf("%f", tx.Value)
	payload += ptypes.TimestampString(tx.Timestamp)
	payload += tx.For
	payload += tx.From
	payload += tx.To

	h := sha256.New()
	h.Write([]byte(payload))
	hash := h.Sum(nil)

	top8 := hash[:8]

	tx.Hash = binary.BigEndian.Uint64(top8)
}
