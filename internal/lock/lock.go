package lock

import (
	"errors"
	"time"
)

var (
	ErrResourceLocked    = errors.New("resource is locked")
	ErrResourceNotLocked = errors.New("resource is not locked")
)

type Locker interface {
	// Lock locks the resource by its name
	// It returns an error if the resource is already locked
	//
	// Example:
	//   Lock("Deployment/my-namespace/my-deployment")
	//
	// This will lock the resource Deployment/my-namespace/my-deployment if it's not already locked
	Lock(name string) error
	// IsLocked checks if the resource is locked
	//
	// Example:
	//   IsLocked("Deployment/my-namespace/my-deployment")
	//
	// This will return true if the resource Deployment/my-namespace/my-deployment is locked
	IsLocked(name string) bool
	// GetLocks returns all locked resources
	//
	// Example:
	//   GetLocks()
	//
	// This will return all locked resources as a slice of strings
	GetLocks() []string
	// Unlock unlocks the resource by its name
	// It returns an error if the resource is not locked
	//
	// Example:
	//   Unlock("Deployment/my-namespace/my-deployment")
	//
	// This will unlock the resource Deployment/my-namespace/my-deployment if it's locked
	Unlock(name string) error
	// ForceUnlockAfter unlocks all resources after the given duration
	//
	// Example:
	//   ForceUnlockAfter(5 * time.Minute)
	//
	// This will unlock all resources after 5 minutes after a lock is acquired
	ForceUnlockAfter(duration time.Duration)
}

type Lock struct {
	locker Locker
}

func NewLock(locker Locker, forceUnlockAfterSec int) *Lock {
	locker.ForceUnlockAfter(time.Duration(forceUnlockAfterSec) * time.Second)

	return &Lock{
		locker: locker,
	}
}

// Lock locks the service by its KindNamespaceName
// It returns an error if the service is already locked
func (l *Lock) Lock(name string) error {
	return l.locker.Lock(name)
}
