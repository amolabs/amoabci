package keys

import (
	"encoding/json"
	"errors"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/xsalsa20symmetric"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/amolabs/amoabci/client/util"
)

func LoadKeyList(path string) (map[string]Key, error) {
	keyList := make(map[string]Key)

	err := util.EnsureFile(path)
	if err != nil {
		return nil, err
	}

	rawKeyList, err := cmn.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(rawKeyList) == 0 {
		return keyList, nil
	}

	err = json.Unmarshal(rawKeyList, &keyList)
	if err != nil {
		return nil, err
	}

	return keyList, nil
}

func SaveKeyList(path string, keyList map[string]Key) error {
	rawKeyList, err := json.Marshal(keyList)
	if err != nil {
		return err
	}

	err = util.EnsureFile(path)
	if err != nil {
		return err
	}

	err = cmn.WriteFile(path, rawKeyList, 0600)
	if err != nil {
		return err
	}

	return nil
}

func Decrypt(key *Key, passphrase []byte) error {
	if !key.Encrypted {
		return errors.New("The key is not encrypted")
	}

	tmp, err := xsalsa20symmetric.DecryptSymmetric(key.PrivKey, crypto.Sha256(passphrase))
	if err != nil {
		return err
	}

	key.Encrypted = false
	key.PrivKey = tmp

	return nil
}
