package nodes

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"google.golang.org/grpc"
)

func (n *node) periodicDiscoverPeers(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	n.discoverPeers(ctx)

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
	var wg sync.WaitGroup
	var mutex sync.Mutex

	if n.knownAddrs.Len() <= 0 {
		n.appendAddrs(n.getSeedAddrs())
	}

	peerAddrs := make(map[string]bool, len(n.peers))
	for _, peer := range n.peers {
		peerAddrs[peer.GetServerAddr()] = true
	}

	unknocked := make(map[string]bool)
	for _, door := range n.knownAddrs.ReadAll() {
		if _, ok := peerAddrs[door]; !ok {
			unknocked[door] = true
		}
	}

	wg.Add(len(unknocked))
	for door := range unknocked {
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
				ServerPort: int32(n.serverPort),
				KnownAddrs: n.getKnownAddrsExcept([]string{door}),
			})
			if err != nil {
				// log.Printf("no answer from %s: %s", door, err)
				n.knownAddrs.RemoveOne(door)
				conn.Close()
				return
			}

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
				conn.Close()
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
				log.Printf("added new peer: %s", door)
				fmt.Println(n.peers)
			} else {
				conn.Close()
			}

			n.appendAddrs(resp.GetKnownAddrs())
		}(door)
	}

	log.Println("peers:")
	for _, p := range n.peers {
		log.Println(p.GetServerAddr())
	}

	wg.Wait()
	// log.Println("known addrs:", n.knownAddrs.ReadAll())
}
