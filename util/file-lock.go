package util

import (
	"os"
	"syscall"
	"time"
)

type fileLock struct {
	fd uintptr
}

func (this *fileLock) lock() error {
	return syscall.Flock(int(this.fd), syscall.LOCK_EX|syscall.LOCK_NB)
}

func (this *fileLock) unlock() error {
	return syscall.Flock(int(this.fd), syscall.LOCK_UN)
}

func (this *fileLock) tryLock() error {
	for i := 0; i < 3; i++ {
		if err := this.lock(); err == nil {
			return nil
		}
		time.Sleep(time.Second * time.Duration(i+1))
	}
	return this.lock()
}

func CreateLockedFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	f.Close()
	return nil
}

func Lock(path string) (func(), error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	lock := &fileLock{fd: f.Fd()}

	if err := lock.tryLock(); err != nil {
		f.Close()
		return nil, err
	}

	return func() {
		lock.unlock()
		f.Close()
	}, nil
}

func WithFileLock(path string, handle func() error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	lock := &fileLock{fd: f.Fd()}

	return withFileLock(lock, handle)
}

func withFileLock(lock *fileLock, handle func() error) error {
	if err := lock.tryLock(); err != nil {
		return err
	}

	defer lock.unlock()

	return handle()
}
