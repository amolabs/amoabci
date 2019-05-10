package keys

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/xsalsa20symmetric"

	"github.com/amolabs/amoabci/client/util"
	"github.com/amolabs/amoabci/crypto/p256"
)

type KeyRing struct {
	filePath string
	keyList  map[string]Key // just a cache
}

func GetKeyRing(path string) (*KeyRing, error) {
	kr := new(KeyRing)
	kr.filePath = path
	kr.keyList = make(map[string]Key)
	err := kr.Load()
	if err != nil {
		return nil, err
	}
	return kr, nil
}

func (kr *KeyRing) Load() error {
	err := util.EnsureFile(kr.filePath)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadFile(kr.filePath)
	if err != nil {
		return err
	}

	newKeyList := make(map[string]Key)
	if len(b) > 0 {
		err = json.Unmarshal(b, &newKeyList)
		if err != nil {
			return err
		}
	}

	kr.keyList = newKeyList

	return nil
}

func (kr *KeyRing) Save() error {
	err := util.EnsureFile(kr.filePath)
	if err != nil {
		return err
	}

	b, err := json.Marshal(kr.keyList)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(kr.filePath, b, 0600)
	if err != nil {
		return err
	}

	return nil
}

func (kr *KeyRing) GenerateNewKey(username string, passphrase []byte, encrypt bool, seed string) (*Key, error) {
	_, ok := kr.keyList[username]
	if ok {
		return nil, errors.New("Username already exists.")
	}

	var privKey p256.PrivKeyP256
	if len(seed) > 0 {
		privKey = p256.GenPrivKeyFromSecret([]byte(seed))
	} else {
		privKey = p256.GenPrivKey()
	}

	return kr.addNewP256Key(privKey, username, passphrase, encrypt)
}

func (kr *KeyRing) ImportPrivKey(keyBytes []byte,
	username string, passphrase []byte, encrypt bool) (*Key, error) {
	_, ok := kr.keyList[username]
	if ok {
		return nil, errors.New("Username already exists.")
	}

	if len(keyBytes) != p256.PrivKeyP256Size {
		return nil, errors.New("Input private key size mismatch.")
	}
	var privKey p256.PrivKeyP256
	copy(privKey[:], keyBytes)

	return kr.addNewP256Key(privKey, username, passphrase, encrypt)
}

func (kr *KeyRing) GetKey(username string) *Key {
	key, ok := kr.keyList[username]
	if !ok {
		return nil
	}
	return &key
}

func (kr *KeyRing) RemoveKey(username string) error {
	_, ok := kr.keyList[username]
	if !ok {
		return errors.New("Username not found")
	}

	delete(kr.keyList, username)

	return kr.Save()
}

func (kr *KeyRing) PrintKeyList() {
	sortKey := make([]string, 0, len(kr.keyList))
	for k := range kr.keyList {
		sortKey = append(sortKey, k)
	}

	sort.Strings(sortKey)

	fmt.Printf("%3s %-9s %-20s %-3s %-40s\n",
		"#", "username", "type", "enc", "address")

	i := 0
	for _, username := range sortKey {
		i++
		key := kr.keyList[username]

		enc := "x"
		if key.Encrypted {
			enc = "o"
		}
		fmt.Printf("%3d %-9s %-20s %-3s %-40s\n",
			i, username, key.Type, enc, key.Address)
	}
}

func (kr *KeyRing) GetNumKeys() int {
	return len(kr.keyList)
}

func (kr *KeyRing) GetFirstKey() *Key {
	var key *Key = nil
	for _, v := range kr.keyList {
		key = &v
		break
	}
	return key
}

func (kr *KeyRing) addNewP256Key(privKey p256.PrivKeyP256,
	username string, passphrase []byte, encrypt bool) (*Key, error) {
	pubKey, ok := privKey.PubKey().(p256.PubKeyP256)
	if !ok {
		return nil, errors.New("Error when deriving pubkey from privkey.")
	}

	key := new(Key)

	key.Type = p256.PrivKeyAminoName
	key.Address = pubKey.Address().String()
	key.PubKey = pubKey.RawBytes()
	if encrypt {
		key.PrivKey = xsalsa20symmetric.EncryptSymmetric(
			privKey.RawBytes(), crypto.Sha256(passphrase))
	} else {
		key.PrivKey = privKey.RawBytes()
	}
	key.Encrypted = encrypt

	kr.keyList[username] = *key
	err := kr.Save()
	if err != nil {
		return nil, err
	}

	return key, nil
}
