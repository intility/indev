package credentialstore

import "fmt"

type CredentialStore interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
}

type CrossPlatformCredentialStore struct {
	innerStore CredentialStore
}

func NewCrossPlatformCredentialStore() *CrossPlatformCredentialStore {
	var innerStore CredentialStore

	if IsKeyringSupported() {
		innerStore = NewKeyringCredentialStore()
	} else {
		innerStore = NewEncryptedFileCredentialStore()
	}

	return &CrossPlatformCredentialStore{
		innerStore: innerStore,
	}
}

func (c *CrossPlatformCredentialStore) Get(key string) ([]byte, error) {
	value, err := c.innerStore.Get(key)
	if err != nil {
		return nil, fmt.Errorf("could not get credentials from underlying store: %w", err)
	}

	return value, nil
}

func (c *CrossPlatformCredentialStore) Set(key string, value []byte) error {
	if err := c.innerStore.Set(key, value); err != nil {
		return fmt.Errorf("could not set credentials in underlying store: %w", err)
	}

	return nil
}

func IsKeyringSupported() bool {
	return true
}
