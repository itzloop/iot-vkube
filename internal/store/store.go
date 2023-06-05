package store

import (
	"context"
	"github.com/itzloop/iot-vkube/types"
	"github.com/pkg/errors"
	"sync"
)

type Store interface {
	RegisterController(ctx context.Context, controller types.Controller) error
	RegisterDevice(ctx context.Context, controllerName string, device types.Device) error
	GetDevices(ctx context.Context, controllerName string) ([]types.Device, error)
	UpdateDevice(ctx context.Context, controllerName string, device types.Device) error
	GetControllers(ctx context.Context) ([]types.Controller, error)
	GetControllersMap(ctx context.Context) (map[string]types.Controller, error)
	UpdateController(ctx context.Context, controller types.Controller) error
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

	return nil
}

func (l *LocalStoreImpl) GetDevices(ctx context.Context, controllerName string) ([]types.Device, error) {
	c, err := l.GetController(ctx, controllerName)
	if err != nil {
		return nil, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	return c.Devices, nil
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
		return errors.New("device does not exist")
	}

	// remove old device
	controller.Devices = append(controller.Devices[:index], controller.Devices[index+1:]...)
	controller.Devices = append(controller.Devices, device)
	return nil
}

func (l *LocalStoreImpl) GetControllers(ctx context.Context) ([]types.Controller, error) {
	var controllers []types.Controller
	l.db.Range(func(key, value any) bool {
		controllers = append(controllers, value.(types.Controller))
		return true
	})

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
func (l *LocalStoreImpl) UpdateController(ctx context.Context, controller types.Controller) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, loaded := l.db.Load(controller.Name)
	if !loaded {
		return errors.New("controller does not exist")
	}

	l.db.Store(controller.Name, controller)
	return nil
}
