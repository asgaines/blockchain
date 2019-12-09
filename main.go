package main

import (
	"context"
	"flag"
	"fmt"
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

const ascii = `
  ____  __     __             
 / __ \/ /    / /  ___  __ __ 
/ /_/ / _ \  / _ \/ _ \/ // / 
\____/_//_/ /_.__/\___/\_, /_
                      /___/|/ 
    __                 ____              __            __   _                            _          
   / /  ___ _______   /  _/ ___  ___    / /  ___  ___ / /  (_)__  ___    ___  ___  ___  (_)__       
  / _ \/ -_) __/ -_) _/ /  / _ \/ _ \  / _ \/ _ \(_-</ _ \/ / _ \/ _ \  / _ \/ _ \/ _ \/ / _ \_  _  _ 
 /_//_/\__/_/  \__/ /___/  \_, /\___/ /_//_/\_,_/___/_//_/_/_//_/\_, /  \_,_/\_, /\_,_/_/_//_(_)(_)(_)
                          /___/                                 /___/       /___/                   
`

// InitialExpectedHashrate is the seed of how many hashes are possible per second.
// The variable set by it is overridden by real data once it comes through.
//
// Setting it too high could lead to the genesis block solve taking a long time
// before the difficulty is adjusted.
const InitialExpectedHashrate = float64(700000)

func main() {
	var pubkey string
	var poolID int
	var addr string
	var minPeers int
	var maxPeers int
	var targetDurPerBlock time.Duration
	var recalcPeriod int
	var speedArg string
	var numMiners int
	var filesPrefix string

	flag.IntVar(&poolID, "poolid", 0, "The ID for a node within a single miner's pool (nodes with same pubkey).")
	flag.StringVar(&addr, "addr", ":20403", "Address to listen on")
	flag.IntVar(&minPeers, "minpeers", 5, "The minimum number of peers to aim for; any fewer will trigger a peer discovery event")
	flag.IntVar(&maxPeers, "maxpeers", 25, "The maximum number of peers to seed out to")
	flag.DurationVar(&targetDurPerBlock, "targetdur", 10*time.Minute, "The desired amount of time between block mining events; controls the difficulty of the mining")
	flag.IntVar(&recalcPeriod, "recalc", 2016, "How many blocks to solve before recalculating difficulty target")
	flag.StringVar(&speedArg, "speed", "medium", "Speed of hashing, CPU usage. One of low/medium/high/ultra")
	flag.IntVar(&numMiners, "miners", 1, "The number of miners to run, one per CPU thread")
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

	fmt.Println(ascii)

	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatalf("invalid addr: %s", addr)
	}

	portN, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(err)
	}

	hasher := chain.NewHasher()

	filesPrefix = fmt.Sprintf("%s_%dp_%dm", targetDurPerBlock, recalcPeriod, numMiners)

	c := chain.InitChain(hasher, filesPrefix)

	difficulty := InitialExpectedHashrate * targetDurPerBlock.Seconds()

	miners := make([]mining.Miner, 0, numMiners)

	for n := 0; n < numMiners; n++ {
		miners = append(miners, mining.NewMiner(
			n,
			(*chain.Block)(c.LastLink()),
			pubkey,
			difficulty,
			targetDurPerBlock,
			speed,
			hasher,
		))
	}

	node := nodes.NewNode(
		c,
		miners,
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
