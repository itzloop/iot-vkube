package smart_lock

import (
	"fmt"
	"sync"
)

// LockController act as the controller that exists
// in households and controllers multiple locks
type LockController struct {
	name  string
	locks sync.Map
}

func NewLockController(name string) *LockController {
	return &LockController{name: name}
}

func (c *LockController) CreateLock(name string) (*SmartLock, error) {
	v, loaded := c.locks.LoadOrStore(name, NewSmartLock(false))
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
