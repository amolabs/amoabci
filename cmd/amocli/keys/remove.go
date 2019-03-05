package keys

import (
	"errors"

	"github.com/amolabs/tendermint-amo/crypto"
	"github.com/amolabs/tendermint-amo/crypto/xsalsa20symmetric"
)

func Remove(nickname string, passphrase []byte, path string) error {
	keyStatus := Check(nickname, path)
	if keyStatus < Exists {
		return errors.New("The key doesn't exist")
	}

	keyList, err := LoadKeyList(path)
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

	err = SaveKeyList(path, keyList)
	if err != nil {
		return err
	}

	return nil
}
