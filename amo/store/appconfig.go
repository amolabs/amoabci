package store

func (s Store) SetAppConfig(b []byte) error {
	s.set([]byte("config"), b)
	return nil
}

func (s Store) GetAppConfig() []byte {
	// always get config from committed tree
	return s.get([]byte("config"), true)
}
