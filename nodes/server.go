package nodes

import (
	"context"
	"errors"
	"log"

	pb "github.com/asgaines/blockchain/protogo/blockchain"
)

func (n *node) Ping(ctx context.Context, r *pb.PingRequest) (*pb.PingResponse, error) {
	// TODO: add the peer making the ping as a peer of this node
	// TODO: share this node's list of peers to the node making the ping
	log.Printf("Got a ping from %d\n", r.GetId())
	return &pb.PingResponse{Ok: true, Id: n.id}, nil
}

func (n *node) SubmitTx(ctx context.Context, r *pb.SubmitTxRequest) (*pb.SubmitTxResponse, error) {
	if r.Tx.GetValue() <= 0 {
		return nil, errors.New("`value` must not be greater than 0")
	}

	if r.Tx.GetFrom() == "" {
		return nil, errors.New("`from` must not be empty")
	}

	if r.Tx.GetTo() == "" {
		return nil, errors.New("`to` must not be empty")
	}

	n.AddTxToQueue(r.Tx)

	return &pb.SubmitTxResponse{Ok: true}, nil
}
