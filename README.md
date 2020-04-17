# Tendermint ABCI App for AMO blockchain
This document is available in [Korean](README.ko.md) also.

## Introduction
Current implementation of AMO blockchain uses
[Tendermint](https://github.com/tendermint/tendermint) as its base consensus
layer. Tendermint handles P2P connection between blockchain nodes; BFT
consensus process among validator nodes; and RPC framework which serves client
requests. However, Tendermint requires ABCI application to interpret block
contents, i.e. transactions. This ABCI app is a main body of a blockchain
application. ABCI app handles processing of transactions, i.e. state transfer;
abstract blockchain state; validator control; and client query for blockchain
state. This repository holds a collection of codes implementing *Tendermint
ABCI app for AMO blockchain* (`amoabci`) and necessary helper scripts.

## Getting Started 

### Install from pre-built binary
Run the following commands to install pre-built amod:
```bash
wget https://github.com/amolabs/amoabci/releases/download/v<version>/amod-<version>-linux-x86_64.tar.gz
tar -xzf amod-<version>-linux-x86_64.tar.gz
sudo cp ./amod /usr/local/bin/amod
```
Specify `<version>` of `amod`. Check out its [latest
releases](https://github.com/amolabs/amoabci/releases)

### Install from source

#### Prerequisites
To build from source, you need to install the followings:
* [git](https://git-scm.com)
* [make](https://www.gnu.org/software/make/)
  * For Debian or Ubuntu linux, you can install `build-essential` package.
  * For MacOS, you can use `make` from Xcode, or install GNU Make via
	[Homebrew](https://brew.sh).
* [golang](https://golang.org/dl/)
  * In some cases, you need to set `GOPATH` and `GOBIN` environment variables
	manually. Check these variables before you proceed.

If you want to run daemons in a docker container or execute some tests
requiring docker, you need install the following:
* [docker](https://www.docker.com) (In Debian or Ubuntu, install `docker.io`)
* [docker-compose](https://www.docker.com)

#### Install amod
Run the following commands to install amod:
```bash
mkdir -p $GOPATH/src/github.com/amolabs
cd $GOPATH/src/github.com/amolabs
git clone https://github.com/amolabs/amoabci
cd amoabci
make install
```

### Install amocli
You can run necessary daemons without `amocli`, but you may want to peek into
blockchain node daemons to see what's going on there. AMO Labs provides a
reference implementation of AMO client(`amocli`) and you may install it to
communicate with AMO blockchain nodes.
```bash
mkdir -p $GOPATH/src/github.com/amolabs
cd $GOPATH/src/github.com/amolabs
git clone https://github.com/amolabs/amo-client-go
cd amo-client-go
make install
```

See [amo-client-go](https://github.com/amolabs/amo-client-go) for more
information.

## Prepare for launch
### Get network information
AMO blockchain node is a networked application. It does nothing useful if not
connected to other nodes. The first thing to figure out is to find out
addresses of other nodes in a AMO blockchain network. Among various nodes in
the network, it is recommended to connect to one of **seed** nodes. If there is
no appropriate seed node, connect to a node having enough **peers**.

* Mainnet information: http://mainnet.amolabs.io
* Testnet information: http://testnet.amolabs.io

*For information about launching a local testnet, see TBA.*

### Get genesis.json
A blockchain is an ever-changing [state
machine](https://en.wikipedia.org/wiki/Finite-state_machine). So you need to
find out what is the initial state of the blockchain. Since AMO blockchain uses
tendermint-like scheme, you need to get `genesis.json` file that defines the
initial state of the chain.

* Mainnet information: http://mainnet.amolabs.io
* Testnet information: http://testnet.amolabs.io

### Prepare data directory
`amod` needs a data directory where they keep configuration file and internal
databases of `amod`. The directory defines a complete snapshot of an AMO
blockchain. So, it is recommended to a keep directory structure something like
the following:
```
(node_data_root)
└── amo 
    ├── config
    └── data
```

`dataroot/amo/config` directory stores some sensitive files such as
`node_key.json` and `priv_validator_key.json`. You need to keep these files
secure by control read permission of them. **Note that this applies to the case
when you run daemons using a docker container**.

### Prepare necessary files
`amod` needs several files to operate properly:
- `config.toml`<sup>&dagger;</sup>: configuration
- `genesis.json`<sup>&dagger;</sup>: initial blockchain and app state
- `node_key.json`<sup>&dagger;&dagger;</sup>: node key for p2p connection
- `priv_validator_key.json`<sup>&dagger;&dagger;</sup>: validator key for
  conesnsus process

&dagger; These files must be prepared before launching `amod`.
Some notable configuration options are as follows:
- `moniker`
- `rpc.laddr`
- `rpc.cors_allowed_origins`
- `p2p.laddr`
- `p2p.external_adderess`
- `p2p.seeds`
- `p2p.persistent_peers`

For more information, see [Tendermint
document](https://tendermint.com/docs/tendermint-core/configuration.html).

&dagger;&dagger; `amod` will generate on its own if not prepared before
launching. But, if you want to use specific keys, of course you need to prepare
it before launching. One possible way to do this is to generate these keys
using `amod tendermint init` command and put them in a configuration directory
along with `config.toml` and `genesis.json`.

## Run initialization
```bash
amod --home <dataroot>/amo tendermint init
```
*NOTE*: To execute tendermint commands, simply append `tendermint` at the end
of `amod`. 

## Run daemon
```bash
amod --home <dataroot>/amo run
```
To run the daemon in background mode, use `amod run &`. Here, `<dataroot>` is a
data directory prepared previously. `amod` will open port 26656 for incoming
P2P connection and port 26657 for incoming RPC connection.

## Run daemons using docker
### Pre-requisites
* [docker](https://www.docker.com) (In Debian or Ubuntu, install `docker.io`)

### Build docker image
You may download the official `amod` docker image(`amolabs/amod`) released from
AMO Labs from [Docker hub](https://hub.docker.com). Of cource, you can build
your own local docker image.

You can build a `amod` docker image by yourself as following or let the `amod`
Makefile do it for you.

To build a `amod` docker image, do the followings:
```bash
mkdir -p $GOPATH/src/github.com/amolabs
cd $GOPATH/src/github.com/amolabs
git clone https://github.com/amolabs/amoabci
cd amoabci
make docker
```
The image will be tagged as `amolabs/amod:latest`. This image includes `amod`,
so you just need one image (and one container).

### Run
Run the daemons in a container as follows:
```bash
docker run -it --rm -p 26656-26657 -v <dataroot>/amo:/amo:Z -d amolabs/amod:latest
```
Options above have the following meaning:
- `-it`: make sure the terminal connects correctly
- `--rm`: remove the container after daemons stop
- `-p 26656-26657`: publish the container's ports to the host machine. This
  make sure that other nodes in the network can connect to our node.
- `-v <dataroot>/amo:/amo:Z`: mount amod data directory.
  **`<dataroot>` must be an absolute path.**
- `amolabs/amod:latest`: use this docker image when creating a container

Make sure that you see series of logs as the daemons init and run.
