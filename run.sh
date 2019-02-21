#!/bin/bash

echo "set moniker =" $MONIKER
echo "set seeds =" $SEEDS

cp config.toml.in config.toml
sed -e s/@moniker@/$MONIKER/ -i.tmp config.toml
sed -e s/@seeds@/$SEEDS/ -i.tmp config.toml
mkdir config
mkdir data
mv -f config.toml config/
if [ "$MONIKER" == "seed" ]; then
	mv node_key.json config/
	mv priv_validator_key.json config/
	mv priv_validator_state.json data/
fi
mv genesis.json config/
amod &
if [ "$MONIKER" == "seed" ]; then
	/usr/bin/tendermint init
fi
/usr/bin/tendermint node
