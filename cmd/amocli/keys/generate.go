package keys

import (
	"github.com/amolabs/tendermint-amo/crypto"
	"github.com/amolabs/tendermint-amo/crypto/p256"
	"github.com/amolabs/tendermint-amo/crypto/xsalsa20symmetric"

	"github.com/amolabs/amoabci/cmd/amocli/util"
)

func Generate(nickname string, passphrase []byte, encrypt bool) error {
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

	keyFile := util.DefaultKeyFilePath()

	keyList, err := LoadKeyList(keyFile)
	if err != nil {
		return err
	}

	keyList[nickname] = newKey

	err = SaveKeyList(keyFile, keyList)
	if err != nil {
		return err
	}

	return nil
}
