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
	"sync"
	"time"

	"github.com/asgaines/blockchain/nodes"
	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"google.golang.org/grpc"
)

func main() {
	var id int
	var ID string
	var port string
	var targetDur time.Duration
	var recalcPeriod int

	// Need a way to assert id is unique across network
	flag.IntVar(&id, "id", 1, "ID for node, should be unique across network")
	flag.StringVar(&ID, "ID", "", "ID for miner running the node; their public key")
	flag.StringVar(&port, "port", ":5050", "Address to listen on")
	flag.DurationVar(&targetDur, "targetdur", 600, "The desired amount of time between block mining events. Used to control the difficulty of the mining.")
	flag.IntVar(&recalcPeriod, "recalc", 2016, "How many blocks to solve before recalculating difficulty target")

	flag.Parse()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	if ascii, err := ioutil.ReadFile("./assets/ascii.txt"); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(string(ascii))
	}

	node := nodes.NewNode(int32(id), ID, targetDur, recalcPeriod)
	wg.Add(1)
	go func() {
		defer wg.Done()
		node.Run(ctx)
	}()

	gsrv := grpc.NewServer()
	pb.RegisterBlockchainServer(gsrv, node)

	lis, err := net.Listen("tcp", port)
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
	signal.Notify(sigs, os.Interrupt)

	go func() {
		<-sigs
		go cancel()
		go func() {
			gsrv.GracefulStop()
		}()
	}()

	wg.Wait()
}
