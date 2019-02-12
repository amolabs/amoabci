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

func (app *AMOApplication) SetAccount(address types.Address, account *types.Account) {
	value, _ := binary.Serialize(account)
	app.state.db.Set(accountFixKey(address[:]), value)
}

func (app *AMOApplication) GetAccount(address types.Address) *types.Account {
	value := app.state.db.Get(accountFixKey(address[:]))
	if len(value) == 0 {
		return &types.Account{
			Balance:        0,
			PurchasedFiles: make(types.HashSet),
		}
	}
	var account types.Account
	err := binary.Deserialize(value, &account)
	if err != nil {
		panic(err)
	}
	return &account
}

func (app *AMOApplication) SetBuyer(fileHash types.Hash, addressSet *types.AddressSet) {
	value, _ := binary.Serialize(addressSet)
	app.state.db.Set(buyerFixKey(fileHash[:]), value)
}

func (app *AMOApplication) GetBuyer(fileHash types.Hash) *types.AddressSet {
	value := app.state.db.Get(buyerFixKey(fileHash[:]))
	if len(value) == 0 {
		return &types.AddressSet{}
	}
	var addressSet types.AddressSet
	err := binary.Deserialize(value, &addressSet)
	if err != nil {
		panic(err)
	}
	return &addressSet
}
