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
	n.appendAddrs(append(r.GetKnownAddrs(), r.NodeID.GetReturnAddr()))

	return &pb.DiscoverResponse{
		Ok:         true, // len(n.peers) < n.maxPeers,
		NodeID:     n.getID().ToProto(),
		KnownAddrs: n.getKnownAddrsExcept([]string{r.NodeID.GetReturnAddr()}),
	}, nil
}

func (n *node) GetState(ctx context.Context, r *pb.GetStateRequest) (*pb.GetStateResponse, error) {
	var c *pb.Chain
	if n.chain != nil {
		c = n.chain.ToProto()
	}

	return &pb.GetStateResponse{
		Chain:      c,
		Difficulty: n.difficulty,
	}, nil
}

func (n *node) ShareChain(ctx context.Context, r *pb.ShareChainRequest) (*pb.ShareChainResponse, error) {
	accepted := n.setChain(&chain.Chain{
		Pbc: r.Chain,
	}, false)

	if accepted {
		n.propagateChain(map[NodeID]bool{
			NodeIDFrom(r.GetNodeID()): true,
		})
	}

	return &pb.ShareChainResponse{Accepted: accepted}, nil
}

func (n *node) ShareTx(ctx context.Context, r *pb.ShareTxRequest) (*pb.ShareTxResponse, error) {
	if r.Tx.GetTimestamp() == nil {
		r.Tx.Timestamp = ptypes.TimestampNow()
	}

	if r.Tx.GetHash() == nil {
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

	for _, miner := range n.miners {
		miner.AddTx(r.Tx)
	}

	var except NodeID
	if nodeID := r.GetNodeID(); nodeID != nil {
		except = NodeIDFrom(nodeID)
	}
	n.propagateTx(r.Tx, except)

	return &pb.ShareTxResponse{Accepted: true}, nil
}

// getPeerAddr is currently a remnant of an attempt to discover requesting peer's
// ip address. Current methods for discovering ip are not reliable within Docker,
// as the ip is reported to the Docker gateway proxy.
// This means connections are reliant upon the requesting peer providing their own correct
// return address
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
