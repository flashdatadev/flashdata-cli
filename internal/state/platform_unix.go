//go:build linux || darwin

package state

import (
	"context"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

func privateCheck(path string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0077 != 0 {
		return ErrUnsafe
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || st.Uid != uint32(os.Geteuid()) {
		return ErrUnsafe
	}
	return nil
}
func privateDirectory(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(path, 0700); err != nil {
			return ErrUnsafe
		}
		info, err = os.Lstat(path)
	}
	if err != nil || !info.IsDir() {
		return ErrUnsafe
	}
	if err = privateCheck(path, info); err != nil {
		return err
	}
	// Resolve parent aliases, while refusing any user-controlled link in the final directory.
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return ErrUnsafe
	}
	if resolved != filepath.Clean(path) {
		return ErrUnsafe
	}
	return nil
}
func secureCreated(path string) error { return os.Chmod(path, 0600) }
func createOutput(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
}
func replace(from, to string) error { return os.Rename(from, to) }
func syncDirectory(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
func lock(ctx context.Context, path string) (func(), error) {
	fd, err := unix.Open(path, unix.O_CREAT|unix.O_RDWR|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
	if err != nil {
		return nil, ErrUnsafe
	}
	f := os.NewFile(uintptr(fd), path)
	info, err := f.Stat()
	if err != nil || privateCheck(path, info) != nil {
		f.Close()
		return nil, ErrUnsafe
	}
	for {
		if err = unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err == nil {
			return func() { _ = unix.Flock(fd, unix.LOCK_UN); f.Close() }, nil
		}
		if err != unix.EWOULDBLOCK && err != unix.EAGAIN {
			f.Close()
			return nil, ErrUnsafe
		}
		select {
		case <-ctx.Done():
			f.Close()
			return nil, ErrBusy
		case <-time.After(25 * time.Millisecond):
		}
	}
}
func protectOS([]byte, bool) ([]byte, error) { return nil, ErrVault }
