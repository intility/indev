package completion

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/adrg/xdg"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/intility/indev/pkg/clientset"
)

type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	return c.now
}

const testHost = "https://prod.example.com"

func newTestCache() (*FileNameCache, *fakeClock) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}

	return &FileNameCache{
		fs:       afero.NewMemMapFs(),
		dir:      "/cache/completion",
		ttl:      time.Minute,
		now:      clock.Now,
		disabled: false,
	}, clock
}

func newTestKey(resource, accountID string) clientset.NameCacheKey {
	return clientset.NameCacheKey{Resource: resource, AccountID: accountID, Host: testHost}
}

func TestFileNameCacheSetGet(t *testing.T) {
	cache, _ := newTestCache()

	cache.Set(newTestKey("clusters", "user-oid.tenant"), []string{"prod-east", "staging"})

	names, ok := cache.Get(newTestKey("clusters", "user-oid.tenant"))

	assert.True(t, ok)
	assert.Equal(t, []string{"prod-east", "staging"}, names)
}

func TestFileNameCacheGetMissingEntry(t *testing.T) {
	cache, _ := newTestCache()

	_, ok := cache.Get(newTestKey("clusters", "nobody"))

	assert.False(t, ok)
}

func TestFileNameCacheGetExpiredEntry(t *testing.T) {
	cache, clock := newTestCache()

	cache.Set(newTestKey("clusters", "user-oid.tenant"), []string{"prod-east"})
	clock.now = clock.now.Add(2 * time.Minute)

	_, ok := cache.Get(newTestKey("clusters", "user-oid.tenant"))

	assert.False(t, ok)
}

func TestFileNameCacheGetCorruptFile(t *testing.T) {
	cache, _ := newTestCache()

	err := afero.WriteFile(cache.fs, cache.filePath(newTestKey("clusters", "user-oid.tenant")), []byte("not json"), 0o600)
	require.NoError(t, err)

	_, ok := cache.Get(newTestKey("clusters", "user-oid.tenant"))

	assert.False(t, ok)
}

func TestFileNameCachePerAccountIsolation(t *testing.T) {
	cache, _ := newTestCache()

	cache.Set(newTestKey("clusters", "alice-oid.tenant"), []string{"prod-east"})
	cache.Set(newTestKey("clusters", "bob-oid.tenant"), []string{"staging"})

	aliceNames, ok := cache.Get(newTestKey("clusters", "alice-oid.tenant"))
	assert.True(t, ok)
	assert.Equal(t, []string{"prod-east"}, aliceNames)

	bobNames, ok := cache.Get(newTestKey("clusters", "bob-oid.tenant"))
	assert.True(t, ok)
	assert.Equal(t, []string{"staging"}, bobNames)
}

func TestFileNameCachePerResourceIsolation(t *testing.T) {
	cache, _ := newTestCache()

	cache.Set(newTestKey("clusters", "user-oid.tenant"), []string{"prod-east"})
	cache.Set(newTestKey("teams", "user-oid.tenant"), []string{"platform"})

	clusterNames, ok := cache.Get(newTestKey("clusters", "user-oid.tenant"))
	assert.True(t, ok)
	assert.Equal(t, []string{"prod-east"}, clusterNames)

	teamNames, ok := cache.Get(newTestKey("teams", "user-oid.tenant"))
	assert.True(t, ok)
	assert.Equal(t, []string{"platform"}, teamNames)
}

func TestFileNameCacheSetEmptyNamesDeletesEntry(t *testing.T) {
	cache, _ := newTestCache()

	cache.Set(newTestKey("clusters", "user-oid.tenant"), nil)

	assert.False(t, aferoExists(t, cache.fs, cache.filePath(newTestKey("clusters", "user-oid.tenant"))))

	cache.Set(newTestKey("clusters", "user-oid.tenant"), []string{"prod-east"})
	cache.Set(newTestKey("clusters", "user-oid.tenant"), []string{})

	_, ok := cache.Get(newTestKey("clusters", "user-oid.tenant"))

	assert.False(t, ok)
	assert.False(t, aferoExists(t, cache.fs, cache.filePath(newTestKey("clusters", "user-oid.tenant"))))
}

func TestFileNameCachePerHostIsolation(t *testing.T) {
	cache, _ := newTestCache()

	prod := newTestKey("clusters", "user-oid.tenant")
	dev := clientset.NameCacheKey{Resource: "clusters", AccountID: "user-oid.tenant", Host: "http://localhost:8080"}

	cache.Set(prod, []string{"prod-east"})

	names, ok := cache.Get(prod)
	assert.True(t, ok)
	assert.Equal(t, []string{"prod-east"}, names)

	_, ok = cache.Get(dev)
	assert.False(t, ok)

	assert.NotEqual(t, cache.filePath(prod), cache.filePath(dev))
}

func TestFileNameCacheEmptyAccountID(t *testing.T) {
	cache, _ := newTestCache()

	key := newTestKey("clusters", "")
	cache.Set(key, []string{"prod-east"})

	names, ok := cache.Get(key)

	assert.True(t, ok)
	assert.Equal(t, []string{"prod-east"}, names)
}

func TestFileNameCacheSanitizedFileName(t *testing.T) {
	cache, _ := newTestCache()

	key := clientset.NameCacheKey{Resource: "ai-apikeys/foo", AccountID: "oid.tenant/suffix x", Host: testHost}

	name := filepath.Base(cache.filePath(key))

	assert.Regexp(t, `^ai-apikeys_foo-[0-9a-f]{16}\.json$`, name)
}

func TestFileNameCacheDelete(t *testing.T) {
	cache, _ := newTestCache()

	cache.Set(newTestKey("clusters", "user-oid.tenant"), []string{"prod-east"})
	cache.Delete(newTestKey("clusters", "user-oid.tenant"))

	_, ok := cache.Get(newTestKey("clusters", "user-oid.tenant"))

	assert.False(t, ok)

	cache.Delete(newTestKey("clusters", "user-oid.tenant"))
}

func TestFileNameCacheClear(t *testing.T) {
	cache, _ := newTestCache()

	cache.Set(newTestKey("clusters", "user-oid.tenant"), []string{"prod-east"})
	cache.Set(newTestKey("teams", "user-oid.tenant"), []string{"platform"})
	cache.Set(newTestKey("clusters", "other-oid.tenant"), []string{"staging"})

	cache.Clear()

	assert.False(t, aferoExists(t, cache.fs, cache.filePath(newTestKey("clusters", "user-oid.tenant"))))
	assert.False(t, aferoExists(t, cache.fs, cache.filePath(newTestKey("teams", "user-oid.tenant"))))
	assert.False(t, aferoExists(t, cache.fs, cache.filePath(newTestKey("clusters", "other-oid.tenant"))))
}

func TestFileNameCacheClearWhenDisabled(t *testing.T) {
	cache, _ := newTestCache()
	cache.disabled = true

	key := newTestKey("clusters", "user-oid.tenant")
	err := afero.WriteFile(cache.fs, cache.filePath(key), []byte("{}"), 0o600)
	require.NoError(t, err)

	cache.Clear()

	assert.False(t, aferoExists(t, cache.fs, cache.filePath(key)))
}

func TestFileNameCacheWriteLeavesNoTempFiles(t *testing.T) {
	cache, _ := newTestCache()

	cache.Set(newTestKey("clusters", "user-oid.tenant"), []string{"prod-east"})

	files, err := afero.ReadDir(cache.fs, cache.dir)
	require.NoError(t, err)

	assert.Len(t, files, 1)
}

func TestFileNameCacheDisabledMethodsAreNoOps(t *testing.T) {
	cache, _ := newTestCache()
	cache.disabled = true

	cache.Set(newTestKey("clusters", "user-oid.tenant"), []string{"prod-east"})

	_, ok := cache.Get(newTestKey("clusters", "user-oid.tenant"))
	assert.False(t, ok)

	err := afero.WriteFile(cache.fs, cache.filePath(newTestKey("clusters", "user-oid.tenant")), []byte("{}"), 0o600)
	require.NoError(t, err)

	cache.Delete(newTestKey("clusters", "user-oid.tenant"))

	assert.True(t, aferoExists(t, cache.fs, cache.filePath(newTestKey("clusters", "user-oid.tenant"))))
}

func aferoExists(t *testing.T, fs afero.Fs, path string) bool {
	t.Helper()

	exists, err := afero.Exists(fs, path)
	require.NoError(t, err)

	return exists
}

func TestNewNameCache(t *testing.T) {
	cache := NewNameCache()

	assert.False(t, cache.disabled)
	assert.Equal(t, nameCacheTTL, cache.ttl)
	assert.Equal(t, filepath.Join(xdg.CacheHome, "indev", "completion"), cache.dir)
}

func TestNewNameCacheDisabledByEnvVar(t *testing.T) {
	t.Setenv(noCompletionCache, "1")

	cache := NewNameCache()

	assert.True(t, cache.disabled)
}

func TestFileNameCacheOnDisk(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}

	cache := &FileNameCache{
		fs:       afero.NewOsFs(),
		dir:      filepath.Join(t.TempDir(), "completion"),
		ttl:      time.Hour,
		now:      clock.Now,
		disabled: false,
	}

	key := newTestKey("clusters", "user-oid.tenant")

	cache.Set(key, []string{"prod-east"})

	names, ok := cache.Get(key)
	assert.True(t, ok)
	assert.Equal(t, []string{"prod-east"}, names)

	assertDiskPerm(t, cache.dir, 0o700)
	assertDiskPerm(t, cache.filePath(key), 0o600)
	assertDiskEntries(t, cache.dir, 1)

	cache.Set(key, []string{"prod-east", "staging"})

	names, ok = cache.Get(key)
	assert.True(t, ok)
	assert.Equal(t, []string{"prod-east", "staging"}, names)
	assertDiskEntries(t, cache.dir, 1)

	cache.Delete(key)

	_, ok = cache.Get(key)
	assert.False(t, ok)
	assertDiskEntries(t, cache.dir, 0)

	cache.Set(key, []string{"staging"})

	cache.Clear()

	_, err := os.Stat(cache.dir)
	assert.True(t, os.IsNotExist(err))
}

func assertDiskPerm(t *testing.T, path string, want os.FileMode) {
	t.Helper()

	if runtime.GOOS == "windows" {
		return
	}

	info, err := os.Stat(path)
	require.NoError(t, err)

	assert.Equal(t, want, info.Mode().Perm())
}

func assertDiskEntries(t *testing.T, dir string, want int) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	assert.Len(t, entries, want)
}
