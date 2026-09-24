package hostd

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func deviceNames(t *testing.T, store *PushTokenStore) []string {
	devices, err := store.List()
	assert.Nil(t, err)
	names := []string{}
	for _, device := range devices {
		names = append(names, device.Name+"="+device.Token)
	}
	return names
}

func TestPushTokenStore(t *testing.T) {
	dir := t.TempDir()
	store := NewPushTokenStore(dir)
	assert.Equal(t, []string{}, deviceNames(t, store))

	device, err := store.Register("token-a:1", " Phone ")
	assert.Nil(t, err)
	assert.Equal(t, "Phone", device.Name)
	assert.NotZero(t, device.RegisteredAt)
	_, err = store.Register("token-b:1", "Tablet")
	assert.Nil(t, err)
	assert.Equal(t, []string{"Phone=token-a:1", "Tablet=token-b:1"}, deviceNames(t, store))
	stat, err := os.Stat(path.Join(dir, pushTokensFile))
	assert.Nil(t, err)
	assert.Equal(t, os.FileMode(0o600), stat.Mode().Perm())

	// rotated token of the same device
	_, err = store.Register("token-a:2", "Phone")
	assert.Nil(t, err)
	assert.Equal(t, []string{"Tablet=token-b:1", "Phone=token-a:2"}, deviceNames(t, store))
	// same token under a new name
	_, err = store.Register("token-b:1", "Old tablet")
	assert.Nil(t, err)
	assert.Equal(t, []string{"Phone=token-a:2", "Old tablet=token-b:1"}, deviceNames(t, store))

	// a fresh store instance sees the same file
	assert.Equal(t, []string{"Phone=token-a:2", "Old tablet=token-b:1"}, deviceNames(t, NewPushTokenStore(dir)))

	assert.Nil(t, store.Unregister("token-a:2"))
	assert.Nil(t, store.Unregister("unknown"))
	assert.Equal(t, []string{"Old tablet=token-b:1"}, deviceNames(t, store))
}

func TestPushTokenStoreValidation(t *testing.T) {
	store := NewPushTokenStore(t.TempDir())
	invalid := []struct{ token, name string }{
		{"", "Phone"},
		{"token", ""},
		{"token", "   "},
		{"has space", "Phone"},
		{"non-ascii-ą", "Phone"},
		{strings.Repeat("a", maxPushTokenLen+1), "Phone"},
		{"token", strings.Repeat("n", maxDeviceNameLen+1)},
		{"token", "line\nbreak"},
	}
	for _, tc := range invalid {
		_, err := store.Register(tc.token, tc.name)
		assert.ErrorIs(t, err, ErrInvalidPushToken, "%q %q", tc.token, tc.name)
	}
	assert.ErrorIs(t, store.Unregister(""), ErrInvalidPushToken)
	assert.Equal(t, []string{}, deviceNames(t, store))
}

func TestPushTokenStoreCorruptFile(t *testing.T) {
	dir := t.TempDir()
	assert.Nil(t, os.WriteFile(path.Join(dir, pushTokensFile), []byte("not json"), 0o600))
	_, err := NewPushTokenStore(dir).List()
	assert.NotNil(t, err)
}
