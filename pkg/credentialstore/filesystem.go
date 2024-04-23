package credentialstore

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

const (
	cacheFileMode = 0o600
	cacheDirMode  = 0o700
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
	err := store.createCacheDirIfNotExists()
	if err != nil {
		return nil, err
	}

	file, err := store.fs.OpenFile(store.filePath, os.O_RDONLY, cacheFileMode)
	if err != nil {
		// the cache file may not exist yet, so we return an empty byte slice
		// instead to indicate that there is no data in the cache
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
	err := store.createCacheDirIfNotExists()
	if err != nil {
		return err
	}

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

func (store *FilesystemCredentialStore) createCacheDirIfNotExists() error {
	dir := filepath.Dir(store.filePath)

	_, err := store.fs.Stat(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = store.fs.MkdirAll(dir, cacheDirMode)
			if err != nil {
				return fmt.Errorf("could not create cache directory: %w", err)
			}
		}
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
