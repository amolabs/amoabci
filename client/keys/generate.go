package keys

import (
	"errors"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/xsalsa20symmetric"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/crypto/p256"
)

func Generate(nickname string, passphrase []byte, encrypt bool, path string) error {
	keyStatus := Check(nickname, path)
	if keyStatus > NoExists {
		return errors.New("The key already exists")
	}

	var privKeyBytes []byte

	privKey := p256.GenPrivKey()
	pubKey := privKey.PubKey()
	p256PubKey, ok := pubKey.(p256.PubKeyP256)
	if !ok {
		return cmn.NewError("Fail to convert public key to p256 public key")
	}

	address := pubKey.Address()

	newKey := Key{
		Type:    p256.PrivKeyAminoName,
		Address: address.String(),
		PubKey:  p256PubKey.RawBytes(),
	}

	if encrypt {
		privKeyBytes = xsalsa20symmetric.EncryptSymmetric(privKey.RawBytes(), crypto.Sha256(passphrase))
	} else {
		privKeyBytes = privKey.RawBytes()
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
