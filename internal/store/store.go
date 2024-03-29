package store

import (
	"context"
	"github.com/itzloop/iot-vkube/types"
	"github.com/pkg/errors"
	"sync"
)

type Store interface {
	ReadOnlyStore
	RegisterController(ctx context.Context, controller types.Controller) error
	RegisterDevice(ctx context.Context, controllerName string, device types.Device) error
	UpdateDevice(ctx context.Context, controllerName string, device types.Device) error
	UpdateController(ctx context.Context, name string, controller types.Controller) error
	DeleteDevice(ctx context.Context, name string, device types.Device) error
	DeleteController(ctx context.Context, name string) error
}

type ReadOnlyStore interface {
	GetDevice(ctx context.Context, controllerName, deviceName string) (types.Device, error)
	GetDevices(ctx context.Context, controllerName string) ([]types.Device, error)
	GetRangeDevices(ctx context.Context, controllerName string, from, to int64) ([]types.Device, error)
	GetController(ctx context.Context, controllerName string) (types.Controller, error)
	GetControllers(ctx context.Context) ([]types.Controller, error)
	GetControllersMap(ctx context.Context) (map[string]types.Controller, error)
}

type LocalStoreImpl struct {
	db *sync.Map
	mu *sync.Mutex
}

func NewLocalStoreImpl() *LocalStoreImpl {
	return &LocalStoreImpl{
		db: &sync.Map{},
		mu: &sync.Mutex{},
	}
}

func (l *LocalStoreImpl) RegisterController(ctx context.Context, controller types.Controller) error {
	if controller.Devices == nil {
		controller.Devices = []types.Device{}
	}

	_, loaded := l.db.LoadOrStore(controller.Name, controller)
	if loaded {
		return errors.New("controller exists")
	}

	return nil
}

func (l *LocalStoreImpl) RegisterDevice(ctx context.Context, controllerName string, device types.Device) error {
	c, err := l.GetController(ctx, controllerName)
	if err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	c.Devices = append(c.Devices, device)
	return l.updateControllerUnsafe(ctx, c.Name, c)
}

func (l *LocalStoreImpl) GetDevices(ctx context.Context, controllerName string) ([]types.Device, error) {
	//c, err := l.GetController(ctx, controllerName)
	//if err != nil {
	//	return nil, err
	//}
	//
	//l.mu.Lock()
	//defer l.mu.Unlock()
	//
	//if c.Devices == nil {
	//	return []types.Device{}, nil
	//}
	//
	//return c.Devices, nil

	return l.GetRangeDevices(ctx, controllerName, -1, -1)
}

func (l *LocalStoreImpl) GetRangeDevices(ctx context.Context, controllerName string, from, to int64) ([]types.Device, error) {
	c, err := l.GetController(ctx, controllerName)
	if err != nil {
		return nil, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if c.Devices == nil {
		return []types.Device{}, nil
	}

	if to == -1 && from == -1 {
		return c.Devices, nil
	} else if from == -1 {
		return c.Devices[:to], nil
	} else if to == -1 {
		return c.Devices[from:], nil
	}

	if from > to {
		return nil, errors.New("from must be smaller that to")
	}

	if to > int64(len(c.Devices)) {
		to = int64(len(c.Devices))
	}

	if from > int64(len(c.Devices)) {
		return []types.Device{}, nil
	}

	return c.Devices[from:to], nil
}

func (l *LocalStoreImpl) GetDevice(ctx context.Context, controllerName, deviceName string) (types.Device, error) {
	devices, err := l.GetDevices(ctx, controllerName)
	if err != nil {
		return types.Device{}, nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	for _, device := range devices {
		if device.Name == deviceName {
			return device, nil
		}
	}

	return types.Device{}, errors.New("getDevice: device not found")
}

func (l *LocalStoreImpl) DeleteDevice(ctx context.Context, controllerName string, device types.Device) error {
	c, err := l.GetController(ctx, controllerName)
	if err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for i, d := range c.Devices {
		if d.Name == device.Name {
			c.Devices = append(c.Devices[:i], c.Devices[i+1:]...)
			l.db.Store(controllerName, c)
			return nil
		}
	}

	return errors.New("deleteDevice: device not found")
}

func (l *LocalStoreImpl) UpdateDevice(ctx context.Context, controllerName string, device types.Device) error {
	controller, err := l.GetController(ctx, controllerName)
	if err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	var index = -1
	for i, d := range controller.Devices {
		if d.Name == device.Name {
			index = i
			break
		}
	}

	if index == -1 {
		return errors.New("updateDevice: device does not exist")
	}

	// remove old device
	controller.Devices = append(controller.Devices[:index], controller.Devices[index+1:]...)
	controller.Devices = append(controller.Devices, device)
	return l.updateControllerUnsafe(ctx, controllerName, controller)
}

func (l *LocalStoreImpl) GetControllers(ctx context.Context) ([]types.Controller, error) {
	var controllers []types.Controller
	l.db.Range(func(key, value any) bool {
		controllers = append(controllers, value.(types.Controller))
		return true
	})

	if controllers == nil {
		controllers = []types.Controller{}
	}

	return controllers, nil
}

func (l *LocalStoreImpl) GetControllersMap(ctx context.Context) (map[string]types.Controller, error) {
	controllers := map[string]types.Controller{}
	l.db.Range(func(key, value any) bool {
		controllers[key.(string)] = value.(types.Controller)
		return true
	})

	return controllers, nil
}

func (l *LocalStoreImpl) GetController(ctx context.Context, controllerName string) (types.Controller, error) {
	v, loaded := l.db.Load(controllerName)
	if !loaded {
		return types.Controller{}, errors.New("controller does not exist")
	}

	return v.(types.Controller), nil
}
func (l *LocalStoreImpl) UpdateController(ctx context.Context, name string, controller types.Controller) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.updateControllerUnsafe(ctx, name, controller)
}

func (l *LocalStoreImpl) DeleteController(ctx context.Context, name string) error {
	_, loaded := l.db.LoadAndDelete(name)
	if !loaded {
		return errors.New("controller does not exist")
	}

	return nil
}

func (l *LocalStoreImpl) updateControllerUnsafe(ctx context.Context, name string, controller types.Controller) error {
	_, loaded := l.db.Load(name)
	if !loaded {
		return errors.New("controller does not exist")
	}

	if name != controller.Name {
		l.db.Delete(name)
	}

	l.db.Store(controller.Name, controller)
	return nil
}
