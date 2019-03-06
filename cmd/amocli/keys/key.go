package keys

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

func Check(nickname string, path string) KeyStatus {
	keyList, err := LoadKeyList(path)
	if err != nil {
		return Unknown
	}

	key, exists := keyList[nickname]
	if !exists {
		return NoExists
	}

	if !key.Encrypted {
		return Exists
	}

	return Encrypted
}
