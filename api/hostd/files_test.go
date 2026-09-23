package hostd

import (
	"io"
	"math"
	"os"
	"path"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
)

func newTestFileStore(t *testing.T) *FileStore {
	store, err := NewFileStore(t.TempDir(), 0)
	assert.Nil(t, err)
	return store
}

func save(store *FileStore, name string, content string, overwrite bool) error {
	_, err := store.Save(name, strings.NewReader(content), int64(len(content)), overwrite)
	return err
}

func listNames(t *testing.T, store *FileStore) []string {
	files, err := store.List()
	assert.Nil(t, err)
	names := []string{}
	for _, file := range files {
		names = append(names, file.Name)
	}
	return names
}

func readFile(t *testing.T, store *FileStore, name string) (string, string) {
	file, err := store.Open(name)
	assert.Nil(t, err)
	defer file.Close()
	content, err := io.ReadAll(file)
	assert.Nil(t, err)
	return string(content), file.Hash
}

func TestFileStoreSaveAndOpen(t *testing.T) {
	store := newTestFileStore(t)
	assert.Equal(t, []string{}, listNames(t, store))

	info, err := store.Save("notes ą.txt", strings.NewReader("hello"), 5, false)
	assert.Nil(t, err)
	assert.Equal(t, "notes ą.txt", info.Name)
	assert.Equal(t, int64(5), info.Size)
	stat, err := os.Stat(path.Join(store.dir, "notes ą.txt"))
	assert.Nil(t, err)
	assert.Equal(t, os.FileMode(0o660), stat.Mode().Perm())
	assert.Equal(t, []string{"notes ą.txt"}, listNames(t, store))

	content, hash := readFile(t, store, "notes ą.txt")
	assert.Equal(t, "hello", content)
	assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", hash)

	assert.ErrorIs(t, save(store, "notes ą.txt", "other", false), ErrFileExists)
	assert.Nil(t, save(store, "notes ą.txt", "replaced", true))
	// the cached hash of the previous content must not be returned
	content, hash = readFile(t, store, "notes ą.txt")
	assert.Equal(t, "replaced", content)
	assert.NotEqual(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", hash)

	assert.Nil(t, store.Delete("notes ą.txt"))
	assert.ErrorIs(t, store.Delete("notes ą.txt"), ErrFileNotFound)
	_, err = store.Open("notes ą.txt")
	assert.ErrorIs(t, err, ErrFileNotFound)

	// nothing left behind by uploads, including the rejected one
	entries, _ := os.ReadDir(store.dir)
	assert.Empty(t, entries)
}

func TestFileStoreNames(t *testing.T) {
	store := newTestFileStore(t)
	invalid := []string{"", "../outside.txt", ".hidden", "..", "a/b", "a\x00b", "a\nb", strings.Repeat("a", 256)}
	for _, name := range invalid {
		assert.ErrorIs(t, save(store, name, "x", true), ErrInvalidFileName, name)
		_, err := store.Open(name)
		assert.ErrorIs(t, err, ErrInvalidFileName, name)
		assert.ErrorIs(t, store.Delete(name), ErrInvalidFileName, name)
	}
	assert.NoFileExists(t, path.Join(path.Dir(store.dir), "outside.txt"))
}

func TestFileStoreIgnoresForeignEntries(t *testing.T) {
	store := newTestFileStore(t)
	secret := path.Join(t.TempDir(), "secret")
	assert.Nil(t, os.WriteFile(secret, []byte("secret"), 0o600))
	assert.Nil(t, os.Symlink(secret, path.Join(store.dir, "link")))
	assert.Nil(t, os.Mkdir(path.Join(store.dir, "directory"), 0o700))
	assert.Nil(t, os.WriteFile(path.Join(store.dir, ".upload-123"), []byte("partial"), 0o600))

	assert.Equal(t, []string{}, listNames(t, store))
	for _, name := range []string{"link", "directory"} {
		_, err := store.Open(name)
		assert.ErrorIs(t, err, ErrFileNotFound, name)
		assert.ErrorIs(t, store.Delete(name), ErrFileNotFound, name)
	}

	// stale uploads are removed when the store is created
	_, err := NewFileStore(store.dir, 0)
	assert.Nil(t, err)
	assert.NoFileExists(t, path.Join(store.dir, ".upload-123"))
	assert.FileExists(t, secret)
}

func TestFileStoreRejectedUploads(t *testing.T) {
	store := newTestFileStore(t)
	cutShort := io.MultiReader(strings.NewReader("abc"), iotest.ErrReader(io.ErrUnexpectedEOF))
	_, err := store.Save("short.txt", cutShort, -1, false)
	assert.ErrorIs(t, err, ErrIncompleteUpload)

	free, err := store.FreeSpace()
	assert.Nil(t, err)
	store.minFreeSpace = free + 1
	assert.ErrorIs(t, save(store, "a.txt", "x", false), ErrNoSpace)
	// declared size does not fit
	store.minFreeSpace = 0
	_, err = store.Save("a.txt", strings.NewReader("x"), math.MaxInt64, false)
	assert.ErrorIs(t, err, ErrNoSpace)

	entries, _ := os.ReadDir(store.dir)
	assert.Empty(t, entries)
}

func TestToken(t *testing.T) {
	stateDir := t.TempDir()
	_, err := ReadToken(stateDir)
	assert.ErrorIs(t, err, os.ErrNotExist)

	token, err := EnsureToken(stateDir)
	assert.Nil(t, err)
	assert.Len(t, token, 64)
	info, _ := os.Stat(path.Join(stateDir, tokenFile))
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	again, _ := EnsureToken(stateDir)
	assert.Equal(t, token, again)
	read, _ := ReadToken(stateDir)
	assert.Equal(t, token, read)
}
