package tx

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/amolabs/amoabci/client/keys"
)

func GetRawKey(path string) (keys.Key, error) {
	var key = keys.Key{}

	keyList, err := keys.LoadKeyList(path)
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

	err = keys.List(path)
	if err != nil {
		return key, nil
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\nType the nickname corresponding to the key for signing this tx: ")
	nickname, err := reader.ReadString('\n')
	if err != nil {
		return key, err
	}

	nickname = strings.Trim(nickname, "\r\n")

	// check the status of key(NoExists, Exists, Encrypted)
	keyStatus := keys.Check(nickname, path)
	if keyStatus == keys.NoExists {
		return key, errors.New("The key doesn't exist")
	}

	key = keyList[nickname]

	// if key is encrypted, request passphrase to decrpyt it
	if key.Encrypted {
		fmt.Printf("Type passphrase: ")
		passphrase, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return key, err
		}

		err = keys.Decrypt(&key, passphrase)
		if err != nil {
			return key, err
		}
	}

	return key, nil
}
