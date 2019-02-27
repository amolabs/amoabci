package keys

import (
	"errors"

	"github.com/amolabs/tendermint-amo/crypto"
	"github.com/amolabs/tendermint-amo/crypto/p256"
	"github.com/amolabs/tendermint-amo/crypto/xsalsa20symmetric"
)

type KeyInfo struct {
	Type      string `json:"type"`
	Address   []byte `json:"address"`
	PubKey    []byte `json:"pub_key"`
	PrivKey   []byte `json:"priv_key"`
	Encrypted bool   `json:"encrypted"`
}

type KeyStatus int

const (
	Unknown KeyStatus = 1 + iota
	NoExists
	Exists
	Encrypted
)

func GenerateKey(nickname string, passphrase []byte, encrypt bool) error {
	var privKeyBytes []byte

	privKey := p256.GenPrivKey()
	pubKey := privKey.PubKey()
	address := pubKey.Address()

	newKey := KeyInfo{
		Type:    p256.PrivKeyAminoName,
		Address: address.Bytes(),
		PubKey:  pubKey.Bytes(),
	}

	if encrypt {
		privKeyBytes = xsalsa20symmetric.EncryptSymmetric(privKey.Bytes(), crypto.Sha256(passphrase))
	} else {
		privKeyBytes = privKey.Bytes()
	}

	newKey.PrivKey = privKeyBytes
	newKey.Encrypted = encrypt

	keyList, err := LoadKeyList()
	if err != nil {
		return err
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

	if key.Encrypted {
		_, err = xsalsa20symmetric.DecryptSymmetric(key.PrivKey, crypto.Sha256(passphrase))
		if err != nil {
			return err
		}
	}

	delete(keyList, nickname)

	err = SaveKeyList(keyList)
	if err != nil {
		return err
	}

	return nil
}

func CheckKey(nickname string) (KeyStatus, error) {
	keyList, err := LoadKeyList()
	if err != nil {
		return Unknown, err
	}

	key, exists := keyList[nickname]
	if !exists {
		return NoExists, errors.New("The key doesn't exist")
	}

	if !key.Encrypted {
		return Exists, errors.New("The key already exists")
	}

	return Encrypted, errors.New("The key already exists")
}
