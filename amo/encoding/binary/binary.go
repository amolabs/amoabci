package binary

type Serializer interface {
	Serialize() ([]byte, error)
}

type Deserializer interface {
	Deserialize([]byte) error
}

func Serialize(v Serializer) ([]byte, error) {
	return v.Serialize()
}

func Deserialize(data []byte, v Deserializer) error {
	return v.Deserialize(data)
}
