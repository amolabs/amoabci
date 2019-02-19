#!/bin/bash

echo "set moniker =" $MONIKER
echo "set seeds =" $SEEDS

cp config.toml.in config.toml
sed -e s/@moniker@/$MONIKER/ -i.tmp config.toml
sed -e s/@seeds@/$SEEDS/ -i.tmp config.toml
mkdir -p blockchain/config
mkdir -p blockchain/data
mv -f config.toml blockchain/config/
if [ "$MONIKER" == "seed" ]; then
	cp node_key.json blockchain/config/
	cp priv_validator_key.json blockchain/config/
	cp priv_validator_state.json blockchain/data/
fi
amod
