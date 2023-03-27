package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

const (
	readPermissions = 0
	AllReadWrite    = os.FileMode(0666)
	userMultiplier  = 64
	groupMultiplier = 8
)

func Read(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		content = []byte{}
		err = nil
	}
	return content, err
}

func Write(path string, permissions fs.FileMode, content []byte) error {
	f, err := createOpen(path, os.O_APPEND|os.O_RDWR|os.O_CREATE|os.O_TRUNC, permissions)
	if err != nil {
		return err
	}
	if _, err := f.Write(content); err != nil {
		return err
	}
	return f.Close()
}

func WriteOpen(path string, permissions fs.FileMode, content []byte) (*os.File, error) {
	f, err := createOpen(path, os.O_APPEND|os.O_RDWR|os.O_CREATE, permissions)
	if err != nil {
		return nil, err
	}
	if _, err := f.Write(content); err != nil {
		return nil, err
	}
	return f, nil
}

func Open(path string) (*os.File, error) {
	f, err := createOpen(path, os.O_APPEND|os.O_RDWR, readPermissions)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// CreateOpen creates or opens the named file. If the file already exists,
// it is opend. If the file does not exist, it is created with mode 0666.
// If successful, methods on the returned File can
// be used for I/O; the associated file descriptor has mode O_RDWR.
func CreateOpen(path string) (*os.File, error) {
	f, err := createOpen(path, os.O_RDWR|os.O_CREATE, AllReadWrite)
	if err != nil {
		return nil, err
	}
	return f, nil
}

var createOpen func(path string, flag int, permissions fs.FileMode) (*os.File, error) = co

func co(path string, flag int, permissions fs.FileMode) (*os.File, error) {
	unix.Umask(0)
	if err := os.MkdirAll(filepath.Dir(path), addExecutible(permissions)); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, flag, permissions)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func addExecutible(permissions fs.FileMode) fs.FileMode {
	user, group, other := calcPermissions(permissions)
	if user%2 == 0 {
		user++
	}
	if group%2 == 0 {
		group++
	}
	if other%2 == 0 {
		other++
	}
	return fileMode(user, group, other)
}

func calcPermissions(permissions fs.FileMode) (user, group, other uint32) {
	permNumber := uint32(permissions)
	return permNumber / userMultiplier,
		permNumber % userMultiplier / groupMultiplier,
		permNumber % userMultiplier % groupMultiplier
}

func fileMode(user, group, other uint32) fs.FileMode {
	return fs.FileMode(user*64 + group*8 + other)
}
