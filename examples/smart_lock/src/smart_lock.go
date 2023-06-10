package smart_lock

import "sync"

type SmartLock struct {
	name   string
	mu     *sync.Mutex
	locked bool
}

func NewSmartLock(name string, initial bool) *SmartLock {
	return &SmartLock{
		name:   name,
		mu:     &sync.Mutex{},
		locked: initial,
	}
}

func (l *SmartLock) Name() string {
	return l.name
}

func (l *SmartLock) Lock() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.locked = true
	return nil
}

func (l *SmartLock) UnLock() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.locked = false
	return nil
}

func (l *SmartLock) Locked() (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	locked := l.locked
	return locked, nil
}

func (l *SmartLock) Readiness() (bool, error) {
	return true, nil
}
