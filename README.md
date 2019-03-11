# Tendermint ABCI App for AMO blockchain

<!--
***NOTE: Tendermint node and the app are built into one single binary in current implementation. This may change in the future.***
-->

## Installation
### Pre-requisites
* [golang](https://golang.org/dl/)
* [golang/dep](https://golang.github.io/dep/docs/installation.html)
* [tendermint-amo](https://github.com/amolabs/tendermint-amo)

### Build from source
* run commands to build Tendermint-amo node:
```bash
git clone https://github.com/amolabs/tendermint-amo
cd tendermint-amo
make get_tools
make get_vendor_deps
make install
```

* run commands to install AMO ABCI app (amod, amocli):
```bash
git clone https://github.com/amolabs/amoabci
cd amoabci
make get_tools
make get_vendor_deps
make install
```
In order to build for another platform (cross-compile) use `TARGET` variable. ex)
```bash
make TARGET=linux install
```

### Gather network information
* mainnet or testnet node address &rarr; $HOME/config/config.toml
* genesis.json &rarr; $HOME/config/genesis.json

### Prepare node
* run commands:
```bash
tendermint init
```

### Run daemons
* First, run ABCI app via the following command:
```bash
amod run
```
**NOTE:** To run the daemon in background mode, run `amod run &`.

* Run Tendermint node via the following command:
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
# If not the first build, get_tools and get_vendor_deps targets are optional.
make get_tools
make get_vendor_deps
make build-linux
make build-docker
```
This will put an image with the tag amolabs/tendermint-amo:latest in the local image pool.

Next, build an amod image
```bash
cd $GOPATH/src/github.com/amolabs/amoabci
# If not the first build, get_tools and get_vendor_deps targets are optional.
make get_tools
make get_vendor_deps
make docker
```
This will put an image with the tag amolabs/amod:latest in the local image pool.

### Run
Run test containers with docker-compose via the following command:
```bash
make run-cluster
```
This will run one seed node and two non-seed validator nodes in *detatched mode*. To run nodes with `stdout` logging, run instead:
```bash
docker-compose up
```

To send a test transaction, run (TODO: fix it):
```bash
docker exec val2 amocli tx transfer --from MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkw --to YTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkw --amount 0
```
And make sure that you see series of logs as the transaction propagate across the nodes and commited in the blockchain.
