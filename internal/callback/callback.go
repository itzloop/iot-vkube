package callback

import (
	"context"
	"github.com/itzloop/iot-vkube/internal/utils"
	"github.com/itzloop/iot-vkube/types"
)

type Callback interface {
	RegisterCallbacks(cb *ServiceCallBacks)
}

type ServiceCallBacks struct {
	OnNewController      func(ctx context.Context, controller types.Controller) error
	OnMissingController  func(ctx context.Context, controller types.Controller) error
	OnExistingController func(ctx context.Context, controller types.Controller) error
	OnNewDevice          func(ctx context.Context, controllerName string, device types.Device) error
	OnMissingDevice      func(ctx context.Context, controllerName string, device types.Device) error
	OnExistingDevice     func(ctx context.Context, controllerName string, device types.Device) error
	OnDeviceDeleted      func(ctx context.Context, controllerName string, device types.Device) error
}

func DefaultServiceCallBacks() *ServiceCallBacks {
	return &ServiceCallBacks{
		OnNewController:      DefaultOnNewController,
		OnMissingController:  DefaultOnMissingController,
		OnExistingController: DefaultOnExistingController,
		OnNewDevice:          DefaultOnNewDevice,
		OnMissingDevice:      DefaultOnMissingDevice,
		OnExistingDevice:     DefaultOnExistingDevice,
	}
}

func DefaultOnNewController(ctx context.Context, controller types.Controller) error {
	spot := "DefaultOnNewController"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Info("invoking callback")
	return nil
}
func DefaultOnMissingController(ctx context.Context, controller types.Controller) error {
	spot := "DefaultOnMissingController"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Info("invoking callback")
	return nil
}
func DefaultOnExistingController(ctx context.Context, controller types.Controller) error {
	spot := "DefaultOnExistingController"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Info("invoking callback")
	return nil
}
func DefaultOnNewDevice(ctx context.Context, controllerName string, device types.Device) error {
	spot := "DefaultOnNewDevice"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Info("invoking callback")
	return nil
}
func DefaultOnMissingDevice(ctx context.Context, controllerName string, device types.Device) error {
	spot := "DefaultOnMissingDevice"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Info("invoking callback")
	return nil
}
func DefaultOnExistingDevice(ctx context.Context, controllerName string, device types.Device) error {
	spot := "DefaultOnExistingDevice"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Info("invoking callback")
	return nil
}
