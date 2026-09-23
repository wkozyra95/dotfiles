package hostd

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"os"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	// Uploads that would leave less free space than this are refused.
	DefaultMinFreeSpace = 5 << 30

	uploadPrefix = ".upload-"
	maxNameLen   = 255
)

var (
	ErrInvalidFileName  = errors.New("invalid file name")
	ErrFileNotFound     = errors.New("file not found")
	ErrFileExists       = errors.New("file already exists")
	ErrNoSpace          = errors.New("not enough free space")
	ErrIncompleteUpload = errors.New("incomplete upload")
)

type FileInfo struct {
	Name     string
	Size     int64
	Modified time.Time
}

// File is an open file from the store, positioned at the start.
type File struct {
	*os.File
	Info FileInfo
	// hex sha256 of the content
	Hash string
}

// FileStore is one flat directory of files to transfer. Names starting with a
// dot are never listed or served, uploads in progress use them.
type FileStore struct {
	dir          string
	minFreeSpace uint64

	// Hashes are cached, because a file has to be read twice otherwise: the
	// hash is needed before the first byte is sent.
	mutex  sync.Mutex
	hashes map[string]cachedHash
}

type cachedHash struct {
	size     int64
	modified time.Time
	hash     string
}

// NewFileStore also cleans up after uploads interrupted by a restart.
func NewFileStore(dir string, minFreeSpace uint64) (*FileStore, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), uploadPrefix) {
			if err := os.Remove(path.Join(dir, entry.Name())); err != nil {
				return nil, err
			}
		}
	}
	return &FileStore{dir: dir, minFreeSpace: minFreeSpace, hashes: map[string]cachedHash{}}, nil
}

func newFileInfo(info os.FileInfo) FileInfo {
	return FileInfo{Name: info.Name(), Size: info.Size(), Modified: info.ModTime()}
}

// filePath is the only way a name received from a client turns into a path.
func (s *FileStore) filePath(name string) (string, error) {
	isInvalid := name == "" || len(name) > maxNameLen || !utf8.ValidString(name) ||
		strings.HasPrefix(name, ".") || strings.ContainsRune(name, '/') ||
		strings.ContainsFunc(name, unicode.IsControl)
	if isInvalid {
		return "", ErrInvalidFileName
	}
	return path.Join(s.dir, name), nil
}

func (s *FileStore) FreeSpace() (uint64, error) {
	stat := syscall.Statfs_t{}
	if err := syscall.Statfs(s.dir, &stat); err != nil {
		return 0, err
	}
	return stat.Bavail * uint64(stat.Bsize), nil
}

func (s *FileStore) List() ([]FileInfo, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	files := []FileInfo{}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		files = append(files, newFileInfo(info))
	}
	return files, nil
}

// Save stores content until EOF, the file shows up under its name only after
// all of it was written. The reader has to fail if the content is cut short.
// sizeHint is the size declared up front (negative if unknown), it only lets
// uploads that can not fit fail early.
func (s *FileStore) Save(name string, content io.Reader, sizeHint int64, overwrite bool) (FileInfo, error) {
	target, err := s.filePath(name)
	if err != nil {
		return FileInfo{}, err
	}
	if _, err := os.Lstat(target); err == nil && !overwrite {
		return FileInfo{}, ErrFileExists
	}
	free, err := s.FreeSpace()
	if err != nil {
		return FileInfo{}, err
	}
	if free < s.minFreeSpace {
		return FileInfo{}, ErrNoSpace
	}
	limit := int64(min(free-s.minFreeSpace, math.MaxInt64-1))
	if sizeHint > limit {
		return FileInfo{}, ErrNoSpace
	}

	upload, err := os.CreateTemp(s.dir, uploadPrefix+"*")
	if err != nil {
		return FileInfo{}, err
	}
	defer func() {
		upload.Close()
		// already gone if it was renamed to the target
		_ = os.Remove(upload.Name())
	}()
	written, err := io.Copy(upload, io.LimitReader(content, limit+1))
	if err != nil {
		log.Warnf("Upload of %s interrupted after %d bytes [%v]", name, written, err)
		return FileInfo{}, ErrIncompleteUpload
	}
	if written > limit {
		return FileInfo{}, ErrNoSpace
	}
	// readable and removable by the group, see nix/nix-modules/hostd.nix
	if err := errors.Join(upload.Chmod(0o660), upload.Sync()); err != nil {
		return FileInfo{}, err
	}

	if overwrite {
		err = os.Rename(upload.Name(), target)
	} else {
		// unlike rename, link does not replace a file that showed up in the meantime
		err = os.Link(upload.Name(), target)
	}
	if errors.Is(err, os.ErrExist) {
		return FileInfo{}, ErrFileExists
	} else if err != nil {
		return FileInfo{}, err
	}
	s.forgetHash(name)
	info, err := os.Lstat(target)
	if err != nil {
		return FileInfo{}, err
	}
	return newFileInfo(info), nil
}

// Open refuses symlinks and anything else that is not a plain file.
func (s *FileStore) Open(name string) (*File, error) {
	filePath, err := s.filePath(name)
	if err != nil {
		return nil, err
	}
	// checked before opening, so that e.g. a fifo does not block
	if info, err := os.Lstat(filePath); errors.Is(err, os.ErrNotExist) || (err == nil && !info.Mode().IsRegular()) {
		return nil, ErrFileNotFound
	} else if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filePath, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ELOOP) {
		return nil, ErrFileNotFound
	} else if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err == nil && !info.Mode().IsRegular() {
		err = ErrFileNotFound
	}
	var hash string
	if err == nil {
		hash, err = s.hash(file, info)
	}
	if err != nil {
		file.Close()
		return nil, err
	}
	return &File{File: file, Info: newFileInfo(info), Hash: hash}, nil
}

func (s *FileStore) Delete(name string) error {
	filePath, err := s.filePath(name)
	if err != nil {
		return err
	}
	if info, err := os.Lstat(filePath); err != nil || !info.Mode().IsRegular() {
		return ErrFileNotFound
	}
	if err := os.Remove(filePath); err != nil {
		return err
	}
	s.forgetHash(name)
	return nil
}

// hash leaves the file positioned at the start.
func (s *FileStore) hash(file *os.File, info os.FileInfo) (string, error) {
	s.mutex.Lock()
	cached, isCached := s.hashes[info.Name()]
	s.mutex.Unlock()
	if isCached && cached.size == info.Size() && cached.modified.Equal(info.ModTime()) {
		return cached.hash, nil
	}
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	hash := hex.EncodeToString(hasher.Sum(nil))
	s.mutex.Lock()
	s.hashes[info.Name()] = cachedHash{size: info.Size(), modified: info.ModTime(), hash: hash}
	s.mutex.Unlock()
	return hash, nil
}

func (s *FileStore) forgetHash(name string) {
	s.mutex.Lock()
	delete(s.hashes, name)
	s.mutex.Unlock()
}
