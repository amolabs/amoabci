package keys

import (
	"fmt"
	"sort"

	"github.com/amolabs/amoabci/cmd/amocli/util"
)

func List() error {
	keyFile := util.DefaultKeyFilePath()

	keyList, err := LoadKeyList(keyFile)
	if err != nil {
		return err
	}

	sortKey := make([]string, 0, len(keyList))
	for k := range keyList {
		sortKey = append(sortKey, k)
	}

	sort.Strings(sortKey)

	fmt.Printf("%-3s %-10s %-20s %-40s\n",
		"seq", "nickname", "type", "address")

	i := 0
	for _, nickname := range sortKey {
		i++
		key := keyList[nickname]

		fmt.Printf("%-3d %-10s %-20s %-40s\n",
			i, nickname, key.Type, key.Address)
	}

	return nil
}
