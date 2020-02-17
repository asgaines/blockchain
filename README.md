# Blockchain

Proof-of-concept of the inner workings of a blockchain implementation.

## Installation

Two options: pull or build

### Pull

`docker pull asgaines/blockchain:latest`

### Build

`docker build -t asgaines/blockchain:latest`

## Run Node

`docker run -p 20403:20403 --rm -v ${PWD}/blockchain_storage:/storage -e BLOCKCHAIN_KEY=<your-key> asgaines/blockchain:latest -returnAddr=<your-ip-or-host>:20403 -seedAddrs=<comma-separated-peer-addrs>`

Required:\
`BLOCKCHAIN_KEY` env variable. It is your "private" key (more similar to a password), used for generating your wallet address and for verifying transactions.\
`-returnAddr` is the address (with port) at which your node is accessible to the rest of the network.


Options:

```
  -bindAddr string
    	Local address to bind/listen on (default ":20403")
  -filesprefix string
    	Common prefix for all output files (default "run")
  -maxpeers int
    	The maximum number of peers to seed out to (default 50)
  -miners int
    	The number of concurrent miners to run, one per thread (default 1)
  -minpeers int
    	The minimum number of peers to aim for; any fewer will trigger a peer discovery event (default 25)
  -poolid int
    	The ID for a node within a single miner's pool (nodes with same pubkey).
  -recalc int
    	How many blocks to solve before recalculating difficulty target (default 10)
  -returnAddr string
    	External address (host:port) for peers to return connections
  -seeds string
      Seeding of potential peers for peer discovery. An optional comma-separated list of host/ips with port.
  -speed string
    	Speed of hashing, CPU usage. One of low/medium/high/ultra (default "medium")
  -targetdur duration
    	The desired amount of time between block mining events; controls the difficulty of the mining (default 10s)
```

## Node Client

Interact with the node directly. Uses gRPC Cobra client interface.

`docker run -i --rm --entrypoint="" asgaines/blockchain:latest go run client/main.go node sharetx -s <node-ip>:20403 <<< '{"tx": {"value": <amount-to-transfer>, "sender": "<your-pubkey>", "recipient": "<recipient-pubkey>", "message": "<optional-metadata>"}}'`