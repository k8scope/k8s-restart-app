package lock

import (
	"fmt"
	"sync"
)

type InMem struct {
	rwmu sync.RWMutex
	m    map[string]struct{}
}

func NewInMem() *InMem {
	return &InMem{
		m: make(map[string]struct{}),
	}
}

func (l *InMem) Lock(name string) error {
	l.rwmu.Lock()
	defer l.rwmu.Unlock()
	if _, ok := l.m[name]; ok {
		return fmt.Errorf("%w: %s", ErrResourceLocked, name)
	}
	l.m[name] = struct{}{}
	return nil
}

func (l *InMem) Unlock(name string) error {
	l.rwmu.Lock()
	defer l.rwmu.Unlock()
	if _, ok := l.m[name]; !ok {
		return fmt.Errorf("%w: %s", ErrResourceNotLocked, name)
	}
	delete(l.m, name)
	return nil
}

func (l *InMem) IsLocked(name string) bool {
	l.rwmu.RLock()
	defer l.rwmu.RUnlock()
	_, ok := l.m[name]
	return ok
}

func (l *InMem) GetLocks() []string {
	l.rwmu.RLock()
	defer l.rwmu.RUnlock()
	locks := make([]string, 0, len(l.m))
	for k := range l.m {
		locks = append(locks, k)
	}
	return locks
}
