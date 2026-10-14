package completion

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/spf13/afero"

	"github.com/intility/indev/pkg/clientset"
)

const (
	nameCacheTTL      = 10 * time.Minute
	noCompletionCache = "INDEV_NO_COMPLETION_CACHE"
	cacheDirMode      = 0o700
)

type nameEntry struct {
	Names     []string  `json:"names"`
	FetchedAt time.Time `json:"fetchedAt"`
}

// FileNameCache caches resource names on disk, keyed per resource, account and API host.
type FileNameCache struct {
	fs       afero.Fs
	dir      string
	ttl      time.Duration
	now      func() time.Time
	disabled bool
}

// NewNameCache returns the on-disk cache for name completion.
func NewNameCache() *FileNameCache {
	return &FileNameCache{
		fs:       afero.NewOsFs(),
		dir:      filepath.Join(xdg.CacheHome, "indev", "completion"),
		ttl:      nameCacheTTL,
		now:      time.Now,
		disabled: os.Getenv(noCompletionCache) != "",
	}
}

func (c *FileNameCache) Get(key clientset.NameCacheKey) ([]string, bool) {
	if c.disabled {
		return nil, false
	}

	entry, ok := c.read(key)
	age := c.now().Sub(entry.FetchedAt)

	if !ok || age > c.ttl {
		return nil, false
	}

	return entry.Names, true
}

func (c *FileNameCache) Set(key clientset.NameCacheKey, names []string) {
	if c.disabled {
		return
	}

	if len(names) == 0 {
		c.Delete(key)

		return
	}

	c.write(key, nameEntry{Names: names, FetchedAt: c.now()})
}

func (c *FileNameCache) Delete(key clientset.NameCacheKey) {
	if c.disabled {
		return
	}

	_ = c.fs.Remove(c.filePath(key))
}

func (c *FileNameCache) Clear() {
	_ = c.fs.RemoveAll(c.dir)
}

func (c *FileNameCache) read(key clientset.NameCacheKey) (nameEntry, bool) {
	var entry nameEntry

	data, err := afero.ReadFile(c.fs, c.filePath(key))
	if err != nil {
		return entry, false
	}

	if err := json.Unmarshal(data, &entry); err != nil {
		return entry, false
	}

	return entry, true
}

func (c *FileNameCache) write(key clientset.NameCacheKey, entry nameEntry) {
	if err := c.fs.MkdirAll(c.dir, cacheDirMode); err != nil {
		return
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	file, err := afero.TempFile(c.fs, c.dir, "names-*")
	if err != nil {
		return
	}

	tempName := file.Name()

	defer func() { _ = c.fs.Remove(tempName) }()

	if _, err := file.Write(data); err != nil {
		_ = file.Close()

		return
	}

	if err := file.Close(); err != nil {
		return
	}

	_ = c.fs.Rename(tempName, c.filePath(key))
}

func (c *FileNameCache) filePath(key clientset.NameCacheKey) string {
	name := sanitizeKey(key.Resource) + "-" + keyHash(key) + ".json"

	return filepath.Join(c.dir, name)
}

func keyHash(key clientset.NameCacheKey) string {
	sum := sha256.Sum256([]byte(key.AccountID + "\x00" + key.Host))

	return hex.EncodeToString(sum[:8])
}

func sanitizeKey(part string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '.':
			return r
		default:
			return '_'
		}
	}, part)
}
