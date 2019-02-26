package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

func LoadFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

func SaveFile(data []byte, filePath string) error {
	return ioutil.WriteFile(filePath, data, 0400)
}
