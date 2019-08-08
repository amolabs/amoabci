echo "set moniker = $MONIKER"
echo "set peers = $PEERS"

iface=$(route -n | grep '^0.0.0.0' | awk '{print $8}')
extaddr=$(ip -f inet a show dev $iface | grep '\<inet\>' | head -1 | awk '{print $2}' | awk -F'/' '{print $1}')
echo "set extaddr = $extaddr"

cp config.toml.in config.toml
sed -e s/@moniker@/$MONIKER/ -i.tmp config.toml
sed -e s/@peers@/$PEERS/ -i.tmp config.toml
sed -e s/@external@/tcp:\\/\\/$extaddr:26656/ -i.tmp config.toml

mkdir -p /tendermint/config/
mkdir -p /tendermint/data/

mv -f config.toml /tendermint/config/
mv -f genesis.json /tendermint/config/

# val1 == genesis validator
if [ "$MONIKER" == "val1" ]; then 
    mv -f priv_validator_key.json /tendermint/config
fi

mv -f priv_validator_state.json /tendermint/data

/usr/bin/amod run &

if [ "$MONIKER" == "val1" ]; then
	/usr/bin/tendermint init
fi

/usr/bin/tendermint node
