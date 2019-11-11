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
	ShareChain(c *chain.Chain, nodeID NodeID, serverPort int32) error
	ShareTx(tx *pb.Tx, nodeID NodeID, serverPort int32) error
	GetServerAddr() string
	Close() error
}

func NewPeer(ctx context.Context, serverAddr string, client pb.NodeClient, conn *grpc.ClientConn) Peer {
	return &peer{
		ctx:        ctx,
		serverAddr: serverAddr,
		client:     client,
		conn:       conn,
	}
}

type peer struct {
	ctx        context.Context
	serverAddr string
	client     pb.NodeClient
	conn       *grpc.ClientConn
}

func (p *peer) ShareChain(c *chain.Chain, nodeID NodeID, serverPort int32) error {
	resp, err := p.client.ShareChain(p.ctx, &pb.ShareChainRequest{
		Chain:      c.ToProto(),
		NodeID:     nodeID.ToProto(),
		ServerPort: serverPort,
	})
	if err != nil {
		return err
	}

	if !resp.GetAccepted() {
		log.Println("peer did not accept chain")
	}

	return nil
}

func (p *peer) ShareTx(tx *pb.Tx, nodeID NodeID, serverPort int32) error {
	resp, err := p.client.ShareTx(p.ctx, &pb.ShareTxRequest{
		Tx:         tx,
		NodeID:     nodeID.ToProto(),
		ServerPort: serverPort,
	})
	if err != nil {
		return err
	}

	if !resp.GetAccepted() {
		log.Println("peer did not accept tx")
	}

	return nil
}

func (p *peer) GetServerAddr() string {
	return p.serverAddr
}

func (p *peer) Close() error {
	return p.conn.Close()
}
