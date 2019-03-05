package keys

import (
	"github.com/amolabs/tendermint-amo/crypto"
	"github.com/amolabs/tendermint-amo/crypto/xsalsa20symmetric"

	"github.com/amolabs/amoabci/cmd/amocli/util"
)

func Remove(nickname string, passphrase []byte) error {
	keyFile := util.DefaultKeyFilePath()

	keyList, err := LoadKeyList(keyFile)
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

	err = SaveKeyList(keyFile, keyList)
	if err != nil {
		return err
	}

	return nil
}
