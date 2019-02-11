package amo

import (
	"github.com/amolabs/amoabci/amo/encoding/binary"
	"github.com/amolabs/amoabci/amo/types"
)

var (
	accountPrefixKey   = []byte("accountKey:")
	fileBuyerPrefixKey = []byte("buyerKey:")
)

func accountFixKey(key []byte) []byte {
	return append(accountPrefixKey, key...)
}

func buyerFixKey(key []byte) []byte {
	return append(fileBuyerPrefixKey, key...)
}

func (app *AMOApplication) SetAccount(account *types.Account) {
	value, _ := binary.Serialize(account)
	app.state.db.Set(accountFixKey([]byte((*account).Address)), value)
}

func (app *AMOApplication) GetAccount(key types.Address) types.Account {
	value := app.state.db.Get(accountFixKey([]byte(key)))
	if len(value) == 0 {
		return types.Account{
			Address:        types.Address(key),
			Balance:        0,
			PurchasedFiles: make(types.HashSet),
		}
	}
	var account types.Account
	err := binary.Deserialize(value, &account)
	if err != nil {
		panic(err)
	}
	return account
}

func (app *AMOApplication) SetBuyer(fileHash types.Hash, addressSet *types.AddressSet) {
	value, _ := binary.Serialize(addressSet)
	app.state.db.Set(buyerFixKey(fileHash[:]), value)
}

func (app *AMOApplication) GetBuyer(fileHash types.Hash) types.AddressSet {
	value := app.state.db.Get(buyerFixKey(fileHash[:]))
	addressSet := types.AddressSet{}
	if len(value) == 0 {
		err := binary.Deserialize(value, &addressSet)
		if err != nil {
			panic(err)
		}
	}
	return addressSet
}
