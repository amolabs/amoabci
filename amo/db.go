package amo

import (
	"encoding/json"
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
	value, _ := json.Marshal(account)
	app.state.db.Set(accountFixKey([]byte(string((*account).Address))), value)
}

func (app *AMOApplication) GetAccount(key types.Address) types.Account {
	value := app.state.db.Get(accountFixKey([]byte(key)))
	if value == nil {
		return types.Account{
			Address:        types.Address(key),
			Balance:        0,
			PurchasedFiles: make(types.HashSet),
		}
	}
	var account types.Account
	err := json.Unmarshal(value, &account)
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
	if value != nil {
		err := binary.Deserialize(value, &addressSet)
		if err != nil {
			panic(err)
		}
	}
	return addressSet
}
