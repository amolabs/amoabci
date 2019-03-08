package keys

import (
	"errors"

	"github.com/amolabs/tendermint-amo/crypto"
	"github.com/amolabs/tendermint-amo/crypto/p256"
	"github.com/amolabs/tendermint-amo/crypto/xsalsa20symmetric"
)

func Generate(nickname string, passphrase []byte, encrypt bool, path string) error {
	keyStatus := Check(nickname, path)
	if keyStatus > NoExists {
		return errors.New("The key already exists")
	}

	var privKeyBytes []byte

	privKey := p256.GenPrivKey()
	pubKey := privKey.PubKey()
	address := pubKey.Address()

	newKey := Key{
		Type:    p256.PrivKeyAminoName,
		Address: address.String(),
		PubKey:  pubKey.Bytes(),
	}

	if encrypt {
		privKeyBytes = xsalsa20symmetric.EncryptSymmetric(privKey.Bytes(), crypto.Sha256(passphrase))
	} else {
		privKeyBytes = privKey.Bytes()
	}

	newKey.PrivKey = privKeyBytes
	newKey.Encrypted = encrypt

	keyList, err := LoadKeyList(path)
	if err != nil {
		return err
	}

	keyList[nickname] = newKey

	err = SaveKeyList(path, keyList)
	if err != nil {
		return err
	}

	return nil
}
