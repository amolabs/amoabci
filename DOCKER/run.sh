#!/bin/bash

echo "set moniker = $MONIKER"
echo "set peers = $PEERS"

iface=$(route -n | grep '^0.0.0.0' | awk '{print $8}')
extaddr=$(ip -f inet a show dev $iface | grep '\<inet\>' | head -1 | awk '{print $2}' | awk -F'/' '{print $1}')
echo "set extaddr = $extaddr"

cp config.toml.in config.toml
sed -e s/@moniker@/$MONIKER/ -i.tmp config.toml
sed -e s/@peers@/$PEERS/ -i.tmp config.toml
sed -e s/@external@/tcp:\\/\\/$extaddr:26656/ -i.tmp config.toml
mkdir config
mkdir data
mv -f config.toml config/
if [ "$MONIKER" == "seed" ]; then
	mv node_key.json config/
	mv priv_validator_key.json config/
	mv priv_validator_state.json data/
fi
mv genesis.json config/
/usr/bin/amod run &
if [ "$MONIKER" == "seed" ]; then
	/usr/bin/tendermint init
fi
/usr/bin/tendermint node
