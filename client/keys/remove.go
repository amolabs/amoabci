package keys

import (
	"errors"
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
		err = GetDecryptedKey(&key)
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
