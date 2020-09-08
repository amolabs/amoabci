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

## Installation 

### Install from pre-built binary
Run the following commands to install pre-built amod:
```bash
wget https://github.com/amolabs/amoabci/releases/download/v<version>/amod-<version>-linux-x86_64.tar.gz
tar -xzf amod-<version>-linux-x86_64.tar.gz
sudo cp ./amod /usr/local/bin/amod
```
Specify `<version>` of `amod`. Check out its [latest
releases](https://github.com/amolabs/amoabci/releases)

### Install from Docker image

#### Install `docker`
Refer to [Get Docker](https://docs.docker.com/get-docker/) in Docker's official
document to install `docker` from either pre-built binary or source.

#### Pull `amolabs/amod` image
To pull the official `amod` image from amolabs, execute the following commands:
```bash
sudo docker pull amolabs/amod:<tag>
```

Specify proper `tag` which indicates a specific version of `amod` image. To
pull the latest image, `tag` should be `latest` or can be omitted. For example,
if you would like to pull `1.7.6`, then execute the following commands: 
```bash
sudo docker pull amolabs/amod:1.7.6
```

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
* [leveldb](https://github.com/google/leveldb)
  * For Debian or Ubuntu linux, you can install 'libleveldb-dev' package.
  * In case you use different servers for building and production, install
    `libleveldb1v5` package in the production server.
* [rocksdb](https://github.com/facebook/rocksdb)
  * For Debian or Ubuntu linux, you can install 'librocksdb-dev' package.
  * In case you use different servers for building and production, install
    `librocksdb5.8` package in the production server.

#### Install `amod`
Run the following commands to build and install `amod`:
```bash
mkdir -p $GOPATH/src/github.com/amolabs
cd $GOPATH/src/github.com/amolabs
git clone https://github.com/amolabs/amoabci
cd amoabci
make install_c
```

### Install amocli
You can run necessary daemons without `amocli`, but you may want to peek into
blockchain node daemons to see what's going on there. AMO Labs provides a
reference implementation of AMO client(`amocli`) and you may install it to
communicate with AMO blockchain nodes. See
[amo-client-go](https://github.com/amolabs/amo-client-go) for more information.

## Preparation
AMO blockchain node is a networked application. It does nothing useful if not
connected to other nodes. The first thing to figure out is to find out
addresses of other nodes in a AMO blockchain network. Among various nodes in
the network, it is recommended to connect to one of **seed** nodes. If there is
no appropriate seed node, connect to a node having enough **peers**.

### Network Information (Seed node)
| chain id | `node_id` | `node_ip_addr` | `node_p2p_port` | `node_rpc_port` |
|-|-|-|-|-|
| `amo-cherryblossom-01` | `fbd1cb0741e30308bf7aae562f65e3fd54359573` | `172.104.88.12` | `26656` | `26657` |
| `amo-testnet-200706` | `a944a1fa8259e19a9bac2c2b41d050f04ce50e51` | `172.105.213.114` | `26656` | `26657` |

**NOTE:** Mainnet's chain id is `amo-cherryblossom-01`. The network information
can be modified without advance notice. If you have a trouble in connecting to
any of these nodes, please feel free to submit a new issue to
[Issues](https://github.com/amolabs/amoabci/issues) section.

### Get `genesis.json`
A blockchain is an ever-changing [state
machine](https://en.wikipedia.org/wiki/Finite-state_machine). So you need to
find out what is the initial state of the blockchain. Since AMO blockchain uses
tendermint-like scheme, you need to get `genesis.json` file that defines the
initial state of the chain.

**NOTE:** If you'd like to launch your own chain for any kind of purposes, you'd
rather generate your own version of `genesis.json` file following
tendermint-like scheme than download existing `genesis.json` file.

To download `genesis.json` file, execute the following command:
```bash
sudo apt install -y curl jq
curl <node_ip_addr>:<node_rpc_port>/genesis | jq '.result.genesis' > genesis.json
```

### Prepare data directory
`amod` needs a data directory where configuration file and internal databases
of `amod` are stored. The directory defines a complete snapshot of an AMO
blockchain. So, it is mandatory to keep a directory structure like the
following:
```
(data_root)
└── amo 
    ├── config
    └── data
```

`data_root/amo/config` directory stores some sensitive files such as
`node_key.json` and `priv_validator_key.json`. You need to keep these files
secure by control of read permission. **Note that this applies to the case when
you run daemons using a docker container as well**.

#### Prepare necessary files
`amod` needs several files located under `data_root/amo/config` to operate
properly:
- `config.toml`<sup>&dagger;</sup>: configuration
- `genesis.json`<sup>&dagger;</sup>: initial blockchain and app state
- `node_key.json`<sup>&dagger;&dagger;</sup>: node key for p2p connection
- `priv_validator_key.json`<sup>&dagger;&dagger;</sup>: validator key for
  conesnsus process

&dagger; These files must be prepared before launching `amod`.

Some notable configuration options of `data_root/amo/config/config.toml` are as
follows:
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
along with `config.toml` and `genesis.json`. Also, It is mandatory to write a
proper seed node's `<node_id>@<node_ip_addr>:<node_p2p_port>` to `p2p.seeds`.
For example, if you'd like to connect to mainnet seed node, `p2p.seeds` would
be `fbd1cb0741e30308bf7aae562f65e3fd54359573@172.104.88.12:26656`.

#### Setup snapshot
Before running a node, there are two available options to sync blocks; sync
from genesis block or sync from snapshot. As syncing from genesis block
consumes lots of physical time, we offer snapshot of blocks taken at certain
block height. The offerings are as follows:
| chain id | `preset` | `version` | `db_backend` | `block_height` | size</br>(comp/raw) |
|-|-|-|-|-|-|
| `amo-cherryblossom-01` | `cherryblossom` | `v1.7.5` | `rocksdb` | `6451392` | 56GB / 116GB |
| `amo-cherryblossom-01` | `cherryblossom` | `v1.6.5` | `rocksdb` | `2908399` | 21GB / 50GB |

**NOTE:** Mainnet's chain id is `amo-cherryblossom-01`.

To download and setup the snapshot, execute the following commands:
```bash
sudo wget http://us-east-1.linodeobjects.com/amo-archive/<preset>_<version>_<db_backend>_<block_height>.tar.bz2
sudo tar -xjf <preset>_<version>_<db_backend>_<block_height>.tar.bz2
sudo rm -rf <data_root>/amo/data/
sudo mv amo-data/amo/data/ <data_root>/amo/
```

**NOTE:** The directory structure of files extracted from compressed `*.tar.bz2`
file may differ from each one. Check out whether extracted `data/` directory is
properly placed under `<data_root>/amo/` directory.

For example, if chain id is `amo-cherryblossom-01`, version is `v1.7.5`, db
backend is `rocksdb`, block height is `6451392`, and data root is `/mynode,
then execute the following commands:
```bash
sudo wget http://us-east-1.linodeobjects.com/amo-archive/cherryblossom_v1.7.5_rocksdb_6451392.tar.bz2
sudo tar -xjf cherryblossom_v1.7.5_rocksdb_6451392.tar.bz2
sudo rm -rf /mynode/amo/data/
sudo mv amo-data/amo/data/ /mynode/amo/
```

## Usage

### Initialize node 
```bash
amod --home <data_root>/amo tendermint init
```
**NOTE**: To execute tendermint commands, simply append `tendermint` at the end
of `amod`. 

### Run node 
```bash
amod --home <data_root>/amo run
```
To run the daemon in background mode, use `amod run &`. Here, `<data_root>` is
a data directory prepared previously. `amod` will open port 26656 for incoming
P2P connection and port 26657 for incoming RPC connection.

## Run node using Docker

### Build Docker image
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

### Run Docker container
Run the daemons in a container as follows:
```bash
docker run -it --rm -p 26656-26657 -v <data_root>/amo:/amo:Z -d amolabs/amod:latest
```
Options above have the following meaning:
- `-it`: make sure the terminal connects correctly
- `--rm`: remove the container after daemons stop
- `-p 26656-26657`: publish the container's ports to the host machine. This
  make sure that other nodes in the network can connect to our node.
- `-v <data_root>/amo:/amo:Z`: mount amod data directory.
  **`<data_root>` must be an absolute path.**
- `amolabs/amod:latest`: use this docker image when creating a container

Make sure that you see series of logs as the daemons init and run.
