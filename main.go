package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
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

func main() {
	var poolID int
	var bindAddr string
	var returnAddr string
	var seedAddrsRaw string
	var minPeers int
	var maxPeers int
	var targetDurPerBlock time.Duration
	var recalcPeriod int
	var speedArg string
	var numMiners int
	var filesPrefix string

	flag.IntVar(&poolID, "poolid", 0, "The ID for a node within a single miner's pool (nodes with same pubkey).")
	flag.StringVar(&bindAddr, "bindAddr", ":20403", "Local address to bind/listen on")
	flag.StringVar(&returnAddr, "returnAddr", "", "External address (host:port) for peers to return connections")
	flag.StringVar(&seedAddrsRaw, "seeds", "", "Seeding of potential peers for peer discovery. An optional comma-separated list of host/ips with port.")
	flag.IntVar(&minPeers, "minpeers", 25, "The minimum number of peers to aim for; any fewer will trigger a peer discovery event")
	flag.IntVar(&maxPeers, "maxpeers", 50, "The maximum number of peers to seed out to")
	flag.DurationVar(&targetDurPerBlock, "targetdur", 10*time.Second, "The desired amount of time between block mining events; controls the difficulty of the mining")
	flag.IntVar(&recalcPeriod, "recalc", 10, "How many blocks to solve before recalculating difficulty target")
	flag.StringVar(&speedArg, "speed", "medium", "Speed of hashing, CPU usage. One of low/medium/high/ultra")
	flag.IntVar(&numMiners, "miners", 1, "The number of concurrent miners to run, one per thread")
	flag.StringVar(&filesPrefix, "filesprefix", "run", "Common prefix for all output files")

	flag.Parse()

	key := os.Getenv("BLOCKCHAIN_KEY")
	if key == "" {
		flag.Usage()
		log.Fatal("Please set BLOCKCHAIN_KEY env variable")
	}

	if returnAddr == "" {
		flag.Usage()
		log.Fatal("please include returnAddr (external host:port) for peers to connect back to your node")
	} else if _, _, err := net.SplitHostPort(returnAddr); err != nil {
		flag.Usage()
		log.Fatal("invalid returnAddr")
	}

	speed, err := mining.ToSpeed(speedArg)
	if err != nil {
		flag.Usage()
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	fmt.Println(ascii)

	hb := sha256.Sum256([]byte(key))
	pubkey := hex.EncodeToString(hb[:])
	log.Printf("Your public key is: %s", pubkey)

	if _, _, err := net.SplitHostPort(bindAddr); err != nil {
		log.Fatalf("invalid bindAddr: %s", bindAddr)
	}

	hasher := chain.NewHasher()
	filesPrefix = fmt.Sprintf("%s_%dp_%dm", targetDurPerBlock, recalcPeriod, numMiners)
	miners := make([]mining.Miner, 0, numMiners)

	for n := 0; n < numMiners; n++ {
		miners = append(miners, mining.NewMiner(
			n,
			pubkey,
			targetDurPerBlock,
			speed,
			hasher,
		))
	}

	node := nodes.NewNode(
		miners,
		pubkey,
		poolID,
		minPeers,
		maxPeers,
		targetDurPerBlock,
		recalcPeriod,
		returnAddr,
		strings.Split(seedAddrsRaw, ","),
		speed,
		filesPrefix,
		hasher,
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		node.Run(ctx)
	}()

	<-node.Ready()

	gsrv := grpc.NewServer()
	pb.RegisterNodeServer(gsrv, node)

	lis, err := net.Listen("tcp", bindAddr)
	if err != nil {
		log.Fatal(err)
	}

	wg.Add(1)
	go func() {
		log.Printf("gRPC server listening on %s", bindAddr)
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
