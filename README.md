# Tendermint ABCI App for AMO blockchain

***NOTE: Tendermint node and the app are built into one single binary in current implementation. This may change in the future.***

## Installation
### Pre-requisites
* [golang](https://golang.org/dl/)
* [golang/dep](https://golang.github.io/dep/docs/installation.html)
* [tendermint](https://github.com/tendermint/tendermint)

### Build from source
* <s>run commands to build Tendermint node:</s>
```bash
git clone https://github.com/tendermint/tendermint
make get_tools
make get_vendor_deps
make install
```

* run commands to build AMO ABCI app:
```bash
git clone https://github.com/amolabs/amoabci
cd amoabci
make
```
In order to build for another platform (cross-compile) use `TARGET` variable. ex)
```bash
make TARGET=linux
```

### Gather network information
* mainnet or testnet node address
* chain ID
* ...

### Run ABCI app
* run commands:
```bash
amod
```

### Prepare keys
* run commands:
```bash
tendermint init
```

### <s>Run Tendermint node</s>
* run commands:
```bash
tendermint node
```

## Test with Docker
For test setup details, see [test-env.md](https://github.com/amolabs/docs/blob/master/test-env.md).

### Pre-requisites
* [tendermint-amo](https://github.com/amolabs/tendermint-amo)
* [docker](https://www.docker.com)
* [docker-compose](https://www.docker.com)

### Build
First, we need to build tendermint node image, and use it as a base image when
building an amod image.
```bash
cd $GOPATH/src/github.com/amolabs/tendermint-amo
make build-linux
make build-docker
```
This will put an image with the tag amolabs/tendermint-amo:latest in the local image pool.

Next, build an amod image
```bash
cd $GOPATH/src/github.com/amolabs/amoabci
make docker
```
This will put an image with the tag amod:latest in the local image pool.

### Run
To run test containers using docker-compose, run:
```bash
make run-cluster
```
This will run one seed node and two non-seed validator nodes.
