package store

import (
	"context"
	"github.com/itzloop/iot-vkube/types"
)

type Store interface {
	RegisterController(ctx context.Context, controller types.Controller) error
	RegisterDevice(ctx context.Context, controllerName string, device types.Device) error
	GetDevices(ctx context.Context, controllerName string) ([]types.Device, error)
	GetControllers(ctx context.Context) ([]types.Controller, error)
}

// TODO implemente this
type LocalStore struct {
}
