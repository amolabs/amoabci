package keys

import (
	"fmt"
	"sort"
)

func List() error {
	keyList, err := LoadKeyList()
	if err != nil {
		return err
	}

	sortKey := make([]string, 0, len(keyList))
	for k := range keyList {
		sortKey = append(sortKey, k)
	}
	sort.Strings(sortKey)

	fmt.Printf("%-3s %-10s %-20s %-40s %-65s\n",
		"seq", "nickname", "type", "address", "pubkey")

	i := 0
	for _, nickname := range sortKey {
		i += 1
		key := keyList[nickname]

		fmt.Printf("%-3d %-10s %-20s %-40x %-65x\n",
			i, nickname, key.Type, key.Address, key.PubKey)
	}

	return nil
}
