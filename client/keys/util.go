package keys

import (
	"encoding/json"

	cmn "github.com/amolabs/tendermint-amo/libs/common"

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
