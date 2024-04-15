package credentialstore

type EncryptedFileCredentialStore struct{}

func NewEncryptedFileCredentialStore() *EncryptedFileCredentialStore {
	return &EncryptedFileCredentialStore{}
}

func (c *EncryptedFileCredentialStore) Get(key string) ([]byte, error) {
	return nil, nil
}

func (c *EncryptedFileCredentialStore) Set(key string, value []byte) error {
	return nil
}
