package keys

import (
	"encoding/json"

	"github.com/amolabs/amoabci/cmd/amocli/util"
)

func LoadKeyList() (map[string]KeyInfo, error) {
	keyList := make(map[string]KeyInfo)

	keyDir := util.DefaultKeyPath()
	keyFile := util.DefaultKeyFilePath()

	util.CreateFile(keyDir, keyFile)

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

	keyDir := util.DefaultKeyPath()
	keyFile := util.DefaultKeyFilePath()

	util.CreateFile(keyDir, keyFile)

	err = util.SaveFile(rawKeyList, keyFile)
	if err != nil {
		return err
	}

	return nil
}
