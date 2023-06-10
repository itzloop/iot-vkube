package smart_lock

import (
	"fmt"
	"sync"
)

// LockController act as the controller that exists
// in households and controllers multiple locks
type LockController struct {
	mu        *sync.Mutex
	name      string
	readiness bool
	locks     sync.Map
}

func NewLockController(name string, readiness bool) *LockController {
	return &LockController{mu: &sync.Mutex{}, name: name, readiness: readiness}
}

func (c *LockController) CreateLock(name string) (*SmartLock, error) {
	v, loaded := c.locks.LoadOrStore(name, NewSmartLock(name, false))
	if loaded {
		return nil, fmt.Errorf("lock '%s' already exists", name)
	}

	return v.(*SmartLock), nil
}

func (c *LockController) GetLock(name string) (*SmartLock, error) {
	v, loaded := c.locks.Load(name)
	if !loaded {
		return nil, fmt.Errorf("lock '%s' doesn't exist", name)
	}

	return v.(*SmartLock), nil
}

func (c *LockController) ListLocks() ([]*SmartLock, error) {
	var list []*SmartLock

	c.locks.Range(func(key, value any) bool {
		sl, ok := value.(*SmartLock)
		if !ok {
			return false
		}

		list = append(list, sl)
		return true
	})

	return list, nil
}

func (c *LockController) SetReadiness(readiness bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.readiness = readiness
}

func (c *LockController) GetReadiness() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.readiness
}
