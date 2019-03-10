package keys

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/amolabs/tendermint-amo/crypto"
	"github.com/amolabs/tendermint-amo/crypto/xsalsa20symmetric"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
	"golang.org/x/crypto/ssh/terminal"

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

func GetKeyToSign(path string) (Key, error) {
	var key = Key{}

	keyList, err := LoadKeyList(path)
	if err != nil {
		return key, err
	}

	switch len(keyList) {
	case 0:
		return key, errors.New("Keys are not found on local storage")
	case 1:
		// safe to use for loop to find the first value of map
		for _, value := range keyList {
			key = value
		}
		return key, nil
	}

	err = List(path)
	if err != nil {
		return key, nil
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\nType the nickname corresponding to the key to sign this tx: ")
	nickname, err := reader.ReadString('\n')
	if err != nil {
		return key, err
	}

	nickname = strings.Replace(nickname, "\r\n", "", -1)
	nickname = strings.Trim(nickname, "\n")

	keyStatus := Check(nickname, path)
	switch keyStatus {
	case NoExists:
		return key, errors.New("The key doesn't exist")
	}

	key = keyList[nickname]

	return key, nil
}

func GetDecryptedKey(key *Key) error {
	if !key.Encrypted {
		return errors.New("The key is not encrypted")
	}

	fmt.Printf("Type passphrase: ")
	passphrase, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return err
	}

	tmp, err := xsalsa20symmetric.DecryptSymmetric(key.PrivKey, crypto.Sha256(passphrase))
	if err != nil {
		return err
	}

	key.Encrypted = false
	key.PrivKey = tmp

	return nil
}
