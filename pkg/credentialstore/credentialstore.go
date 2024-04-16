package credentialstore

type CredentialStore interface {
	// Get retrieves the data associated with the given partition key.
	Get(partitionKey string) ([]byte, error)
	// Set stores the given data with the given partition key.
	Set(data []byte, partitionKey string) error
	// Clear removes all data from the store.
	Clear() error
}
