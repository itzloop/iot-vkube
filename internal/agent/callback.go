package agent

import (
	"context"
	"github.com/itzloop/iot-vkube/internal/callback"
	"github.com/itzloop/iot-vkube/types"
	"github.com/itzloop/iot-vkube/utils"
)

func (service *Service) ServiceCallBacks() *callback.ServiceCallBacks {
	return &callback.ServiceCallBacks{
		OnNewController:      service.OnNewController,
		OnMissingController:  service.OnMissingController,
		OnExistingController: service.OnExistingController,
		OnNewDevice:          service.OnNewDevice,
		OnMissingDevice:      service.OnMissingDevice,
		OnExistingDevice:     service.OnExistingDevice,
		OnDeviceDeleted:      service.OnDeviceDeleted,
	}
}

func (service *Service) OnNewController(ctx context.Context, controller types.Controller) error {
	spot := "agent/OnNewController"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Trace("invoking callback")

	// TODO notify controller
	return service.store.RegisterController(ctx, controller)
}
func (service *Service) OnMissingController(ctx context.Context, controller types.Controller) error {
	spot := "agent/OnMissingController"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Trace("invoking callback")
	return nil
}
func (service *Service) OnExistingController(ctx context.Context, controller types.Controller) error {
	spot := "agent/OnExistingController"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Trace("invoking callback")
	return nil
}
func (service *Service) OnNewDevice(ctx context.Context, controllerName string, device types.Device) error {
	spot := "agent/OnNewDevice"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Trace("invoking callback")

	// TODO notify controller
	return service.store.RegisterDevice(ctx, controllerName, device)
}
func (service *Service) OnMissingDevice(ctx context.Context, controllerName string, device types.Device) error {
	spot := "agent/OnMissingDevice"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Trace("invoking callback")
	return nil
}
func (service *Service) OnExistingDevice(ctx context.Context, controllerName string, device types.Device) error {
	spot := "agent/OnExistingDevice"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Trace("invoking callback")
	return nil
}

func (service *Service) OnDeviceDeleted(ctx context.Context, controllerName string, device types.Device) error {
	spot := "agent/OnDeviceDeleted"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Trace("invoking callback")

	// TODO notify controller
	return service.store.DeleteDevice(ctx, controllerName, device)
}
