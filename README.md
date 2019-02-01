# Tendermint ABCI App for AMO blockchain

## Installation
### Pre-requisites
* [golang](https://golang.org/dl/)
* [golang/dep](https://golang.github.io/dep/docs/installation.html)
* [tendermint](https://github.com/tendermint/tendermint)

### Build from source
* run commands to build Tendermint node:
```bash
git clone https://github.com/tendermint/tendermint
make get_tools
make get_vendor_deps
make install
```

* run commands to build AMO ABCI app:
```bash
git clone https://github.com/amolabs/amoabci
dep ensure
go install
```

### Gather network information
* mainnet or testnet node address
* chain ID
* ...

### Run ABCI app
* run commands:
```bash
amoabci
```

### Prepare keys
* run commands:
```bash
tendermint init
```

### Run Tendermint node
* run commands:
```bash
tendermint node
```

