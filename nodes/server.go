package nodes

import (
	"context"
	"errors"
	"log"
	"net"
	"strconv"

	"github.com/asgaines/blockchain/chain"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"github.com/asgaines/blockchain/transactions"
	"github.com/golang/protobuf/ptypes"
	grpcpeer "google.golang.org/grpc/peer"
)

func (n *node) Discover(ctx context.Context, r *pb.DiscoverRequest) (*pb.DiscoverResponse, error) {
	peerAddr := n.getPeerAddr(ctx, r.ServerPort)

	n.appendAddrs(append(r.GetKnownAddrs(), peerAddr))

	return &pb.DiscoverResponse{
		Ok:         true,
		NodeID:     n.getID().ToProto(),
		KnownAddrs: n.getKnownAddrsExcept([]string{peerAddr}),
	}, nil
}

func (n *node) ShareChain(ctx context.Context, r *pb.ShareChainRequest) (*pb.ShareChainResponse, error) {
	accepted := n.setChain(&chain.Chain{
		Pbc: r.Chain,
	}, false)
	//log.Printf("Did I accept peer chain? %v\n", accepted)
	return &pb.ShareChainResponse{Accepted: accepted}, nil
}

func (n *node) ShareTx(ctx context.Context, r *pb.ShareTxRequest) (*pb.ShareTxResponse, error) {
	if r.Tx.GetTimestamp() == nil {
		r.Tx.Timestamp = ptypes.TimestampNow()
	}

	if r.Tx.GetHash() == 0 {
		transactions.SetHash(r.Tx)
	}

	if r.Tx.GetValue() <= 0 {
		return nil, errors.New("`value` must be greater than 0")
	}

	if r.Tx.GetFrom() == "" {
		return nil, errors.New("`from` must not be empty")
	}

	if r.Tx.GetTo() == "" {
		return nil, errors.New("`to` must not be empty")
	}

	n.miner.AddTx(r.Tx)

	var except NodeID
	if nodeID := r.GetNodeID(); nodeID != nil {
		except = NodeIDFrom(nodeID)
	}
	n.propagateTx(r.Tx, except)

	return &pb.ShareTxResponse{Accepted: true}, nil
}

func (n *node) getPeerAddr(ctx context.Context, port int32) string {
	gpeer, ok := grpcpeer.FromContext(ctx)
	if !ok {
		log.Println("could not get address of peer")
	}

	host, _, err := net.SplitHostPort(gpeer.Addr.String())
	if err != nil {
		log.Println(err)
	}

	return net.JoinHostPort(host, strconv.Itoa(int(port)))
}
