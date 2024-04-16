package credentialstore

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/spf13/afero"
)

const (
	cacheFileMode = 0o600
)

type FilesystemCredentialStore struct {
	filePath string
	fs       afero.Fs
}

type FileSystemOption func(*FilesystemCredentialStore)

func NewFilesystemCredentialStore(filePath string, options ...FileSystemOption) *FilesystemCredentialStore {
	store := &FilesystemCredentialStore{
		filePath: filePath,
		fs:       afero.NewOsFs(),
	}

	for _, option := range options {
		option(store)
	}

	return store
}

func WithFilesystem(fs afero.Fs) FileSystemOption {
	return func(store *FilesystemCredentialStore) {
		store.fs = fs
	}
}

func (store *FilesystemCredentialStore) Get(partitionKey string) ([]byte, error) {
	file, err := store.fs.OpenFile(store.filePath, os.O_RDONLY, cacheFileMode)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []byte{}, nil
		}

		return nil, fmt.Errorf("could not read cache file: %w", err)
	}

	cacheData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read cache file: %w", err)
	}

	return cacheData, nil
}

func (store *FilesystemCredentialStore) Set(data []byte, partitionKey string) error {
	file, err := store.fs.OpenFile(store.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, cacheFileMode)
	if err != nil {
		return fmt.Errorf("could not open cache file: %w", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("could not write to cache file: %w", err)
	}

	return nil
}

func (store *FilesystemCredentialStore) Clear() error {
	err := store.fs.Remove(store.filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("could not remove cache file: %w", err)
	}

	return nil
}
