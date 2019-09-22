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
)

func AddLilBits(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var lb LilBits
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&lb); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	node, err := clstr.AddTransaction(lb)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := w.Write(node.GetChain().ToJSON()); err != nil {
		log.Fatal(err)
	}
}

var clstr Cluster

func main() {
	var numNodes int
	var targetMin float64
	flag.IntVar(&numNodes, "numnodes", 10, "Number of nodes to spin up for the blockchain network.")
	flag.Float64Var(&targetMin, "targetminutes", 1, "The target for the lapse between block additions. Used to control the difficulty of the mining.")
	flag.Parse()

	chain := InitBlockchain()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	clstr = NewCluster(numNodes, chain, targetMin)
	wg.Add(1)
	go func() {
		defer wg.Done()
		clstr.Run(ctx)
	}()

	http.Handle("/add/lilbits", http.HandlerFunc(AddLilBits))

	server := http.Server{
		Addr: ":8080",
	}

	sigRec := make(chan os.Signal)
	signal.Notify(sigRec, os.Interrupt)

	go func() {
		<-sigRec
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
