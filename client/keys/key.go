package keys

import (
	"errors"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/xsalsa20symmetric"
)

type Key struct {
	Type      string `json:"type"`
	Address   string `json:"address"`
	PubKey    []byte `json:"pub_key"`
	PrivKey   []byte `json:"priv_key"`
	Encrypted bool   `json:"encrypted"`
}

func (key *Key) Decrypt(passphrase []byte) error {
	if !key.Encrypted {
		return errors.New("The key is not encrypted")
	}

	plainKey, err := xsalsa20symmetric.DecryptSymmetric(key.PrivKey, crypto.Sha256(passphrase))
	if err != nil {
		return err
	}

	key.Encrypted = false
	key.PrivKey = plainKey

	return nil
}
