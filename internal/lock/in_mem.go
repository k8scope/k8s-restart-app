package lock

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type InMem struct {
	rwmu sync.RWMutex
	m    map[string]time.Time
}

func NewInMem() *InMem {
	return &InMem{
		m: make(map[string]time.Time),
	}
}

func (l *InMem) Lock(name string) error {
	l.rwmu.Lock()
	defer l.rwmu.Unlock()
	if _, ok := l.m[name]; ok {
		return fmt.Errorf("%w: %s", ErrResourceLocked, name)
	}
	l.m[name] = time.Now()
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

func (l *InMem) ForceUnlockAfter(duration time.Duration) {
	go func() {
		for {
			for k, v := range l.m {
				if time.Since(v) > duration {
					err := l.Unlock(k)
					if err != nil {
						slog.Error("failed to force unlock resource", "error", err)
						continue
					}
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()
}
