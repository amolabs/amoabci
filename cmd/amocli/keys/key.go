package keys

import (
	"errors"

	"github.com/amolabs/amoabci/cmd/amocli/util"
)

type Key struct {
	Type      string `json:"type"`
	Address   string `json:"address"`
	PubKey    []byte `json:"pub_key"`
	PrivKey   []byte `json:"priv_key"`
	Encrypted bool   `json:"encrypted"`
}

type KeyStatus int

const (
	Unknown KeyStatus = 1 + iota
	NoExists
	Exists
	Encrypted
)

func CheckKey(nickname string) (KeyStatus, error) {
	keyFile := util.DefaultKeyFilePath()

	keyList, err := LoadKeyList(keyFile)
	if err != nil {
		return Unknown, err
	}

	key, exists := keyList[nickname]
	if !exists {
		return NoExists, errors.New("The key doesn't exist")
	}

	if !key.Encrypted {
		return Exists, errors.New("The key already exists")
	}

	return Encrypted, errors.New("The key already exists (encrypted)")
}
