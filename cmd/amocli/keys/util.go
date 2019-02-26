package keys

import (
	"encoding/json"
	"os"

	"github.com/amolabs/amoabci/cmd/amocli/util"
)

func LoadKeyList() (map[string]KeyInfo, error) {
	keyList := make(map[string]KeyInfo)

	keyFile := util.DefaultKeyFilePath()

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		os.MkdirAll(util.DefaultKeyPath(), os.ModePerm)
		os.OpenFile(keyFile, os.O_RDONLY|os.O_CREATE, 0600)
	}

	rawKeyList, err := util.LoadFile(keyFile)
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

func SaveKeyList(keyList map[string]KeyInfo) error {
	rawKeyList, err := json.Marshal(keyList)
	if err != nil {
		return err
	}

	keyFile := util.DefaultKeyFilePath()

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		os.MkdirAll(util.DefaultKeyPath(), os.ModePerm)
		os.OpenFile(keyFile, os.O_RDONLY|os.O_CREATE, 0600)
	}

	err = util.SaveFile(rawKeyList, keyFile)
	if err != nil {
		return err
	}

	return nil
}
