package keys

import (
	"errors"

	"github.com/amolabs/tendermint-amo/crypto"
	"github.com/amolabs/tendermint-amo/crypto/p256"
	"github.com/amolabs/tendermint-amo/crypto/xsalsa20symmetric"
)

type KeyInfo struct {
	Type       string `json:"type"`
	Address    []byte `json:"address"`
	PubKey     []byte `json:"pubKey"`
	EncPrivKey []byte `json:"encPrivKey"`
}

func GenerateKey(nickname string, passphrase []byte) error {
	privKey := p256.GenPrivKey()
	pubKey := privKey.PubKey()
	address := pubKey.Address()

	encPrivKey := xsalsa20symmetric.EncryptSymmetric(privKey.Bytes(), crypto.Sha256(passphrase))

	keyList, err := LoadKeyList()
	if err != nil {
		return err
	}

	newKey := KeyInfo{
		Type:       p256.PrivKeyAminoName,
		Address:    address.Bytes(),
		PubKey:     pubKey.Bytes(),
		EncPrivKey: encPrivKey,
	}

	keyList[nickname] = newKey

	err = SaveKeyList(keyList)
	if err != nil {
		return err
	}

	return nil
}

func RemoveKey(nickname string, passphrase []byte) error {
	keyList, err := LoadKeyList()
	if err != nil {
		return err
	}

	key := keyList[nickname]

	_, err = xsalsa20symmetric.DecryptSymmetric(key.EncPrivKey, crypto.Sha256(passphrase))
	if err != nil {
		return err
	}

	delete(keyList, nickname)

	err = SaveKeyList(keyList)
	if err != nil {
		return err
	}

	return nil
}

func CheckKey(nickname string) (bool, error) {
	keyList, err := LoadKeyList()
	if err != nil {
		return false, err
	}

	_, exists := keyList[nickname]
	if !exists {
		return false, errors.New("The key doesn't exist")
	}

	return true, errors.New("The key already exists")
}
