module github.com/amolabs/amoabci

go 1.13

require (
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.6.3
	github.com/stretchr/testify v1.6.0
	github.com/tendermint/iavl v0.13.3
	github.com/tendermint/tendermint v0.33.5
	github.com/tendermint/tm-db v0.5.1
)

replace github.com/tendermint/iavl => github.com/amolabs/iavl v0.13.3-amo1
