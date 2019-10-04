package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/asgaines/blockchain/cluster"
	"github.com/asgaines/blockchain/transactions"
)

func AddTransaction(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var tx transactions.Tx
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	clstr.AddTransaction(tx)

	if _, err := w.Write([]byte(`{"accepted": true}`)); err != nil {
		log.Fatal(err)
	}
}

var clstr cluster.Cluster

func main() {
	var numNodes int
	var targetDur time.Duration
	var recalcPeriod int

	flag.IntVar(&numNodes, "numnodes", 10, "Number of nodes to spin up for the blockchain network.")
	flag.DurationVar(&targetDur, "targetdur", 600, "The desired amount of time between block mining events. Used to control the difficulty of the mining.")
	flag.IntVar(&recalcPeriod, "recalc", 2016, "How many blocks to solve before recalculating difficulty target")

	flag.Parse()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	clstr = cluster.NewCluster(numNodes, targetDur, recalcPeriod)
	wg.Add(1)
	go func() {
		defer wg.Done()
		clstr.Run(ctx)
	}()

	http.Handle("/add/tx", http.HandlerFunc(AddTransaction))

	server := http.Server{
		Addr: ":8080",
	}

	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt)

	go func() {
		<-sigs
		go cancel()
		go func() {
			if err := server.Shutdown(ctx); err != nil {
				log.Println(err)
			}
		}()
	}()

	wg.Add(1)
	go func() {
		ascii, err := ioutil.ReadFile("./assets/ascii.txt")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(ascii))
		log.Println(server.ListenAndServe())
		wg.Done()
	}()

	wg.Wait()
}
