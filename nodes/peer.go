package nodes

import (
	"context"
	"errors"

	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"google.golang.org/grpc"
)

type Peer interface {
	GetID() int32
	SubmitTx(tx *pb.Tx) error
	Close() error
}

func NewPeer(id int32, client pb.BlockchainClient, conn *grpc.ClientConn) Peer {
	return &peer{
		node: node{
			id: id,
		},
		client: client,
		conn:   conn,
	}
}

type peer struct {
	node   node
	client pb.BlockchainClient
	conn   *grpc.ClientConn
}

func (p *peer) GetID() int32 {
	return p.node.id
}

func (p *peer) SubmitTx(tx *pb.Tx) error {
	req := &pb.SubmitTxRequest{Tx: tx}
	resp, err := p.client.SubmitTx(context.Background(), req)
	if err != nil {
		return err
	}

	if !resp.GetOk() {
		return errors.New("did not receive ok from peer")
	}

	return nil
}

func (p *peer) Close() error {
	return p.conn.Close()
}
