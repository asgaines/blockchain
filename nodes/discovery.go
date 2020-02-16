package nodes

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/asgaines/blockchain/chain"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"google.golang.org/grpc"
)

func (n *node) periodicDiscoverPeers(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if len(n.peers) < n.minPeers {
				n.discoverPeers(ctx)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (n *node) discoverPeers(ctx context.Context) {
	// log.Printf("Discovering peers. Current peers: %v. Known addrs: %v", n.peers, n.knownAddrs)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	if n.knownAddrs.Len() < 1 {
		n.appendAddrs(n.getSeedAddrs())
	}

	wg.Add(n.knownAddrs.Len())
	for _, door := range n.knownAddrs.ReadAll() {
		go func(door string) {
			defer wg.Done()

			conn, err := grpc.Dial(door, grpc.WithInsecure())
			if err != nil {
				log.Printf("could not dial address %s: %s", door, err)
				n.knownAddrs.RemoveOne(door)
				return
			}

			client := pb.NewNodeClient(conn)

			resp, err := client.Discover(ctx, &pb.DiscoverRequest{
				NodeID:     n.getID().ToProto(),
				KnownAddrs: n.getKnownAddrsExcept([]string{door}),
			})
			if err != nil {
				n.knownAddrs.RemoveOne(door)
				conn.Close()
				return
			}

			// log.Printf("received known addrs: %v", resp.GetKnownAddrs())

			nodeID := NodeIDFrom(resp.GetNodeID())
			if nodeID == n.getID() {
				n.knownAddrs.RemoveOne(door)
				conn.Close()
				return
			}

			// Having to check this is indicative of an issue
			// This door ideally would not have been in the list to begin with
			// It's due to knownAddrs having two different addresses for the same node,
			// one likely from a seed, the other from the peer reaching out
			if _, ok := n.peers[nodeID]; ok {
				// n.knownAddrs.RemoveOne(door)
				// log.Println(n.knownAddrs.ReadAll())
				if err := conn.Close(); err != nil {
					log.Println(err)
				}
				return
			}

			if resp.GetOk() && len(n.peers) < n.maxPeers {
				mutex.Lock()
				n.peers[nodeID] = NewPeer(
					ctx,
					door,
					client,
					conn,
				)
				mutex.Unlock()
				log.Printf("Added new peer: %s (address: %s)", nodeID.ToProto().GetPubkey(), door)
				// fmt.Println(n.peers)
			} else {
				if err := conn.Close(); err != nil {
					log.Println(err)
				}
			}

			n.appendAddrs(resp.GetKnownAddrs())
		}(door)
	}

	wg.Wait()
}

func (n *node) getInitState(ctx context.Context) (*chain.Chain, float64, error) {
	mainChain := chain.InitChain(n.hasher, n.filesPrefix)
	difficulty := InitialExpectedHashrate * n.targetDurPerBlock.Seconds()

	var wg sync.WaitGroup
	var mutex sync.Mutex

	wg.Add(len(n.peers))
	for _, p := range n.peers {
		go func(p Peer) {
			defer wg.Done()

			c, diff, err := p.GetState(n.getID())
			if err != nil {
				log.Println(err)
				return
			}

			if !n.IsValid(c) {
				return
			}

			if c.Length() > mainChain.Length() {
				mutex.Lock()
				mainChain = c
				difficulty = diff
				mutex.Unlock()
			}
		}(p)
	}

	wg.Wait()

	return mainChain, difficulty, nil
}
