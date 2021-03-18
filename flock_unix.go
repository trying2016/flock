// Copyright 2015 Tim Heckman. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

// +build !aix,!windows

package flock

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

func (f *Flock) LockRead() error {
	return f.lock(false, false, 0, os.SEEK_SET, 0)
}

func (f *Flock) LockWrite() error {
	return f.lock(true, false, 0, os.SEEK_SET, 0)
}

func (f *Flock) LockReadB() error {
	return f.lock(false, true, 0, os.SEEK_SET, 0)
}

func (f *Flock) LockWriteB() error {
	return f.lock(true, true, 0, os.SEEK_SET, 0)
}

func (f *Flock) Unlock() error {
	f.unlock(0, os.SEEK_SET, 0)
	return nil
}

func (f *Flock) LockReadRange(offset int64, whence int, len int64) error {
	return f.lock(false, false, offset, whence, len)
}

func (f *Flock) LockWriteRange(offset int64, whence int, len int64) error {
	return f.lock(true, false, offset, whence, len)
}

func (f *Flock) LockReadRangeB(offset int64, whence int, len int64) error {
	return f.lock(false, true, offset, whence, len)
}

func (f *Flock) LockWriteRangeB(offset int64, whence int, len int64) error {
	return f.lock(true, true, offset, whence, len)
}

func (f *Flock) UnlockRange(offset int64, whence int, len int64) {
	f.unlock(offset, whence, len)
}

// Owner will return the pid of the process that owns an fcntl based
// lock on the file. If the file is not locked it will return -1. If
// a lock is owned by the current process, it will return -1.
func (f *Flock) Owner() int {
	ft := &syscall.Flock_t{}
	*ft = *f.getFt()

	err := syscall.FcntlFlock(f.fh.Fd(), syscall.F_GETLK, ft)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	if ft.Type == syscall.F_UNLCK {
		fmt.Println(err)
		return -1
	}

	return int(ft.Pid)
}

func (f *Flock) lock(exclusive, blocking bool, offset int64, whence int, len int64) error {
	if f.fh == nil {
		file, err := os.OpenFile(f.Path(), os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			return err
		}
		f.fh = file
	}

	ft := &syscall.Flock_t{
		Whence: int16(whence),
		Start:  offset,
		Len:    len,
		Pid:    int32(os.Getpid()),
	}
	f.ft = ft

	if exclusive {
		ft.Type = syscall.F_WRLCK
	} else {
		ft.Type = syscall.F_RDLCK
	}
	var flags int
	if blocking {
		flags = syscall.F_SETLKW
	} else {
		flags = syscall.F_SETLK
	}

	err := syscall.FcntlFlock(f.fh.Fd(), flags, f.getFt())
	if err != nil {
		return errors.New("err failed to lock")
	}

	return nil
}

func (f *Flock) unlock(offset int64, whence int, len int64) {
	f.getFt().Len = len
	f.getFt().Start = offset
	f.getFt().Whence = int16(whence)
	f.getFt().Type = syscall.F_UNLCK
	syscall.FcntlFlock(f.fh.Fd(), syscall.F_SETLK, f.getFt())
}

func (f *Flock) getFt() *syscall.Flock_t {
	return f.ft.(*syscall.Flock_t)
}
