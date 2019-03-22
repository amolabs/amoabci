package keys

import (
	"errors"

	"github.com/amolabs/amoabci/crypto/p256"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/xsalsa20symmetric"
	cmn "github.com/tendermint/tendermint/libs/common"
)

func Import(rawPrivKey []byte, nickname string, passphrase []byte, encrypt bool, path string) error {
	keyStatus := Check(nickname, path)
	if keyStatus > NoExists {
		return errors.New("The key with nickname '" + nickname + "' already exists")
	}

	var privKey p256.PrivKeyP256
	copy(privKey[:], rawPrivKey)

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
		rawPrivKey = xsalsa20symmetric.EncryptSymmetric(privKey.RawBytes(), crypto.Sha256(passphrase))
	} else {
		rawPrivKey = privKey.RawBytes()
	}

	newKey.PrivKey = rawPrivKey
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
