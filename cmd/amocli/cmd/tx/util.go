package tx

import (
	"errors"

	"github.com/amolabs/amoabci/client/keys"
	"github.com/amolabs/amoabci/cmd/amocli/util"
)

var (
	Username   string
	Passphrase string
	UserKey    keys.Key
)

func GetRawKey(path string) (keys.Key, error) {
	empty := keys.Key{}
	kr, err := keys.GetKeyRing(path)
	if err != nil {
		return empty, err
	}

	switch kr.GetNumKeys() {
	case 0:
		return empty, errors.New("Empty key ring.")
	case 1:
		return *kr.GetFirstKey(), nil
	}

	if len(Username) == 0 {
		kr.PrintKeyList()
		Username, err = util.PromptUsername()
		if err != nil {
			return empty, err
		}
	}

	key := kr.GetKey(Username)
	if key == nil {
		return empty, errors.New("Key not found")
	}

	// if key is encrypted, request passphrase to decrpyt it
	if key.Encrypted {
		if len(Passphrase) == 0 {
			Passphrase, err = util.PromptPassphrase()
			if err != nil {
				return *key, err
			}
		}

		err = key.Decrypt([]byte(Passphrase))
		if err != nil {
			return *key, err
		}
	}

	return *key, nil
}
