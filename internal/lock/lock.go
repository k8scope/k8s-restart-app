package lock

import (
	"errors"
	"log/slog"
	"time"
)

var (
	ErrResourceLocked    = errors.New("resource is locked")
	ErrResourceNotLocked = errors.New("resource is not locked")
)

type Locker interface {
	Lock(name string) error
	IsLocked(name string) bool
	GetLocks() []string
	Unlock(name string) error
}

type Lock struct {
	locker Locker
}

func NewLock(locker Locker, forceUnlockAfterSec int) *Lock {
	go func() {
		for {
			// force unlock all resources
			for _, name := range locker.GetLocks() {
				err := locker.Unlock(name)
				if err != nil {
					slog.Error("failed to force unlock resource", "error", err)
				}
			}
			time.Sleep(time.Duration(forceUnlockAfterSec) * time.Second)
		}
	}()

	return &Lock{
		locker: locker,
	}
}

// Lock locks the service by its KindNamespaceName
// It returns an error if the service is already locked
func (l *Lock) Lock(name string) error {
	return l.locker.Lock(name)
}
