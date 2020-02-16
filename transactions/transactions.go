package transactions

import (
	"crypto/sha256"
	"fmt"

	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"github.com/golang/protobuf/ptypes"
)

func SetHash(tx *pb.Tx) {
	payload := fmt.Sprintf("%f", tx.GetValue())
	payload += ptypes.TimestampString(tx.GetTimestamp())
	payload += tx.GetSender()
	payload += tx.GetRecipient()
	payload += tx.GetMessage()

	h := sha256.New()
	h.Write([]byte(payload))
	tx.Hash = h.Sum(nil)
}
