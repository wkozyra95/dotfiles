package hostd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	pushTokensFile = "push-tokens.json"

	maxPushTokenLen  = 4096
	maxDeviceNameLen = 64
)

var ErrInvalidPushToken = errors.New("invalid push token")

// PushDevice is a mobile device that registered its FCM token, `mycli mobile
// send` delivers push notifications to it.
type PushDevice struct {
	Token string `json:"token"`
	// Chosen by the device, unique among the registered devices.
	Name string `json:"name"`
	// unix seconds
	RegisteredAt int64 `json:"registeredAt"`
}

type pushTokensFileContent struct {
	Devices []PushDevice `json:"devices"`
}

// PushTokenStore keeps the registered devices in a JSON file in the state
// directory. The server writes it, `mycli mobile send` reads it (and removes
// tokens FCM reports as gone), so every operation reads the file fresh and
// writes it atomically.
type PushTokenStore struct {
	file  string
	mutex sync.Mutex
}

func NewPushTokenStore(stateDir string) *PushTokenStore {
	return &PushTokenStore{file: path.Join(stateDir, pushTokensFile)}
}

func (s *PushTokenStore) load() ([]PushDevice, error) {
	content, err := os.ReadFile(s.file)
	if errors.Is(err, os.ErrNotExist) {
		return []PushDevice{}, nil
	}
	if err != nil {
		return nil, err
	}
	parsed := pushTokensFileContent{}
	if err := json.Unmarshal(content, &parsed); err != nil {
		return nil, fmt.Errorf("%s: %w", s.file, err)
	}
	if parsed.Devices == nil {
		parsed.Devices = []PushDevice{}
	}
	return parsed.Devices, nil
}

func (s *PushTokenStore) save(devices []PushDevice) error {
	content, err := json.MarshalIndent(pushTokensFileContent{Devices: devices}, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.file + ".tmp"
	if err := os.WriteFile(tmp, append(content, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.file)
}

// List returns the devices in the order they were registered.
func (s *PushTokenStore) List() ([]PushDevice, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.load()
}

// Register adds a device or updates it: a device that re-registers with a new
// token (FCM rotates them) replaces its old entry by name, a token that moved
// to a new name replaces the old one by token.
func (s *PushTokenStore) Register(token string, name string) (PushDevice, error) {
	name = strings.TrimSpace(name)
	if err := validatePushToken(token); err != nil {
		return PushDevice{}, err
	}
	if err := validateDeviceName(name); err != nil {
		return PushDevice{}, err
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	devices, err := s.load()
	if err != nil {
		return PushDevice{}, err
	}
	device := PushDevice{Token: token, Name: name, RegisteredAt: time.Now().Unix()}
	kept := []PushDevice{}
	for _, existing := range devices {
		if existing.Token != token && existing.Name != name {
			kept = append(kept, existing)
		}
	}
	return device, s.save(append(kept, device))
}

// Unregister removes the device with that token, it is not an error if there is none.
func (s *PushTokenStore) Unregister(token string) error {
	if err := validatePushToken(token); err != nil {
		return err
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	devices, err := s.load()
	if err != nil {
		return err
	}
	kept := []PushDevice{}
	for _, existing := range devices {
		if existing.Token != token {
			kept = append(kept, existing)
		}
	}
	if len(kept) == len(devices) {
		return nil
	}
	return s.save(kept)
}

func validatePushToken(token string) error {
	if token == "" {
		return fmt.Errorf("%w: token is required", ErrInvalidPushToken)
	}
	if len(token) > maxPushTokenLen {
		return fmt.Errorf("%w: token is too long", ErrInvalidPushToken)
	}
	for _, c := range token {
		if c <= ' ' || c > '~' {
			return fmt.Errorf("%w: token contains unexpected characters", ErrInvalidPushToken)
		}
	}
	return nil
}

func validateDeviceName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: device name is required", ErrInvalidPushToken)
	}
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) > maxDeviceNameLen {
		return fmt.Errorf("%w: device name is too long or not valid UTF-8", ErrInvalidPushToken)
	}
	for _, c := range name {
		if unicode.IsControl(c) {
			return fmt.Errorf("%w: device name contains control characters", ErrInvalidPushToken)
		}
	}
	return nil
}
