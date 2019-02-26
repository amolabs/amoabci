package keys

import (
	"fmt"
)

func List() error {
	keyList, err := LoadKeyList()
	if err != nil {
		return err
	}

	fmt.Printf("%-3s %-10s %-20s %-40s %-65s\n",
		"seq", "nickname", "type", "address", "pubkey")

	i := 0
	for nickname, key := range keyList {
		i += 1
		fmt.Printf("%-3d %-10s %-20s %-40x %-65x\n",
			i, nickname, key.Type, key.Address, key.PubKey)
	}

	return nil
}
