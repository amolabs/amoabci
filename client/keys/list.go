package keys

import (
	"fmt"
	"sort"
)

func List(path string) error {
	keyList, err := LoadKeyList(path)
	if err != nil {
		return err
	}

	sortKey := make([]string, 0, len(keyList))
	for k := range keyList {
		sortKey = append(sortKey, k)
	}

	sort.Strings(sortKey)

	fmt.Printf("%-3s %-10s %-20s %-10s %-40s\n",
		"seq", "nickname", "type", "encrypted", "address")

	i := 0
	for _, nickname := range sortKey {
		i++
		key := keyList[nickname]

		fmt.Printf("%-3d %-10s %-20s %-10t %-40s\n",
			i, nickname, key.Type, key.Encrypted, key.Address)
	}

	return nil
}
