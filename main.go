package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/asgaines/blockchain/chain"
	"github.com/asgaines/blockchain/mining"
	"github.com/asgaines/blockchain/nodes"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"google.golang.org/grpc"
)

// InitialExpectedHashrate is the seed of how many hashes are possible per second.
// The variable set by it is overridden by real data once it comes through.
//
// Setting it too high could lead to the genesis block solve taking a long time
// before the difficulty is adjusted.
const InitialExpectedHashrate = float64(100)

func main() {
	var pubkey string
	var poolID int
	var addr string
	var minPeers int
	var maxPeers int
	var targetDurPerBlock time.Duration
	var recalcPeriod int
	var speedArg string
	var filesPrefix string

	flag.IntVar(&poolID, "poolid", 0, "The ID for a node within a single miner's pool (nodes with same pubkey).")
	flag.StringVar(&addr, "addr", ":20403", "Address to listen on")
	flag.IntVar(&minPeers, "minpeers", 5, "The minimum number of peers to aim for; any fewer will trigger a peer discovery event")
	flag.IntVar(&maxPeers, "maxpeers", 25, "The maximum number of peers to seed out to")
	flag.DurationVar(&targetDurPerBlock, "targetdur", 10*time.Minute, "The desired amount of time between block mining events; controls the difficulty of the mining")
	flag.IntVar(&recalcPeriod, "recalc", 2016, "How many blocks to solve before recalculating difficulty target")
	flag.StringVar(&speedArg, "speed", "medium", "Speed of hashing, CPU usage. One of low/medium/high/ultra")
	flag.StringVar(&filesPrefix, "filesprefix", "run", "Common prefix for all output files")

	flag.Parse()

	pubkey = os.Getenv("BLOCKCHAIN_PUBKEY")

	if pubkey == "" {
		flag.Usage()
		log.Fatal("pubkey missing from environment. Please set $BLOCKCHAIN_PUBKEY env variable.")
	}

	speed, err := mining.ToSpeed(speedArg)
	if err != nil {
		flag.Usage()
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	if ascii, err := ioutil.ReadFile("./assets/ascii.txt"); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(string(ascii))
	}

	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatalf("invalid addr: %s", addr)
	}

	portN, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(err)
	}

	hasher := chain.NewHasher()

	c := chain.InitChain(hasher, filesPrefix)

	difficulty := InitialExpectedHashrate * targetDurPerBlock.Seconds()

	miner := mining.NewMiner(
		(*chain.Block)(c.LastLink()),
		pubkey,
		difficulty,
		targetDurPerBlock,
		speed,
		nodes.CalcTarget(difficulty),
		hasher,
	)

	node := nodes.NewNode(
		c,
		miner,
		pubkey,
		poolID,
		minPeers,
		maxPeers,
		targetDurPerBlock,
		recalcPeriod,
		portN,
		speed,
		filesPrefix,
		difficulty,
		hasher,
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		node.Run(ctx)
	}()

	gsrv := grpc.NewServer()
	pb.RegisterNodeServer(gsrv, node)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	wg.Add(1)
	go func() {
		log.Printf("gRPC server listening on %s", port)
		if err := gsrv.Serve(lis); err != nil {
			log.Printf("gRPC server: %v", err)
		}
		wg.Done()
	}()
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		go cancel()
		go func() {
			gsrv.GracefulStop()
		}()
	}()

	wg.Wait()
}
