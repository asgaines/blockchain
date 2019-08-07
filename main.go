package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

func AddLilBits(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	chain := r.Context().Value(chainKey).(*Blockchain)

	var lb LilBits
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&lb); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	chain.addBlock(lb)

	log.Printf("%#v\n", lb)

	if i, err := w.Write(chain.ToJSON()); err != nil {
		log.Println(i, err)
	}
}

func WithChain(next http.Handler, chain *Blockchain) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), chainKey, chain)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

type ctxkey int

const chainKey ctxkey = iota

func main() {
	chain := InitBlockchain()

	http.Handle("/add/lilbits", WithChain(http.HandlerFunc(AddLilBits), &chain))

	log.Println("Hollah at ya server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
