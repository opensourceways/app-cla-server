package util

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type FileLock struct {
	path string
	f    *os.File
}

func (this *FileLock) Lock() error {
	f, err := os.Open(this.path)
	if err != nil {
		return err
	}

	this.f = f

	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func (this *FileLock) Unlock() error {
	defer this.f.Close()

	return syscall.Flock(int(this.f.Fd()), syscall.LOCK_UN)
}

func (this *FileLock) CreateLockedFile() error {
	f, err := os.OpenFile(this.path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	f.Close()
	return nil
}

func NewFileLock(path string) *FileLock {
	return &FileLock{path: path}
}

func LockedFilePath(dir, platform, org, repo string) string {
	s := filepath.Join(platform, org, repo)
	return filepath.Join(dir, strings.ReplaceAll(s, string(filepath.Separator), "_"))
}
