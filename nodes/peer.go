package nodes

import (
	"context"
	"log"

	"github.com/asgaines/blockchain/chain"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"google.golang.org/grpc"
)

// Peer manages a client connection to a Node running at a different address
type Peer interface {
	ShareChain(c *chain.Chain, nodeID NodeID, returnAddr string) error
	ShareTx(tx *pb.Tx, nodeID NodeID, returnAddr string) error
	Close() error
}

func NewPeer(ctx context.Context, returnAddr string, client pb.NodeClient, conn *grpc.ClientConn) Peer {
	return &peer{
		ctx:        ctx,
		returnAddr: returnAddr,
		client:     client,
		conn:       conn,
	}
}

type peer struct {
	ctx        context.Context
	returnAddr string
	client     pb.NodeClient
	conn       *grpc.ClientConn
}

func (p *peer) ShareChain(c *chain.Chain, nodeID NodeID, returnAddr string) error {
	resp, err := p.client.ShareChain(p.ctx, &pb.ShareChainRequest{
		Chain:      c.ToProto(),
		NodeID:     nodeID.ToProto(),
		ReturnAddr: returnAddr,
	})
	if err != nil {
		return err
	}

	if !resp.GetAccepted() {
		log.Println("PEER DID NOT ACCEPT CHAIN")
	} else {
		log.Println("PEER ACCEPTED CHAIN")
	}

	return nil
}

func (p *peer) ShareTx(tx *pb.Tx, nodeID NodeID, returnAddr string) error {
	resp, err := p.client.ShareTx(p.ctx, &pb.ShareTxRequest{
		Tx:         tx,
		NodeID:     nodeID.ToProto(),
		ReturnAddr: returnAddr,
	})
	if err != nil {
		return err
	}

	if !resp.GetAccepted() {
		log.Println("peer did not accept tx")
	}

	return nil
}

func (p *peer) Close() error {
	return p.conn.Close()
}
