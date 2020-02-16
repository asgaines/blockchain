package nodes

import (
	pb "github.com/asgaines/blockchain/protogo/blockchain"
)

func NodeIDFrom(nodeID *pb.NodeID) NodeID {
	return NodeID{
		Pubkey: nodeID.Pubkey,
		Id:     nodeID.Id,
	}
}

type NodeID struct {
	Pubkey     string
	Id         int32
	ReturnAddr string
}

func (n NodeID) ToProto() *pb.NodeID {
	return &pb.NodeID{
		Pubkey:     n.Pubkey,
		Id:         n.Id,
		ReturnAddr: n.ReturnAddr,
	}
}

func (n NodeID) IsSamePool(nodeID NodeID) bool {
	return n.Pubkey == nodeID.Pubkey
}
