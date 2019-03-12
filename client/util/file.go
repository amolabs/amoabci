package util

import (
	"os"
	"path/filepath"

	cmn "github.com/tendermint/tendermint/libs/common"
)

var (
	defaultCLIDir      = ".amocli"
	defaultKeyDir      = "keys"
	defaultKeyListFile = "keys.json"
)

func defaultCLIPath() string {
	return filepath.Join(os.ExpandEnv("$HOME"), defaultCLIDir)
}

func DefaultKeyPath() string {
	return filepath.Join(defaultCLIPath(), defaultKeyDir)
}

func DefaultKeyFilePath() string {
	return filepath.Join(DefaultKeyPath(), defaultKeyListFile)
}

func EnsureFile(path string) error {
	dirPath, _ := filepath.Split(path)

	err := cmn.EnsureDir(dirPath, 0775)
	if err != nil {
		return err
	}

	_, err = os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	return err
}
