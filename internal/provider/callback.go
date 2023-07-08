package provider

import (
	"context"
	"github.com/itzloop/iot-vkube/internal/callback"
	"github.com/itzloop/iot-vkube/types"
	"github.com/itzloop/iot-vkube/utils"
)

func (p *PodLifecycleHandlerImpl) ServiceCallBacks() *callback.ServiceCallBacks {
	return &callback.ServiceCallBacks{
		OnNewController:      p.OnNewController,
		OnMissingController:  p.OnMissingController,
		OnExistingController: p.OnExistingController,
		OnNewDevice:          p.OnNewDevice,
		OnMissingDevice:      p.OnMissingDevice,
		OnExistingDevice:     p.OnExistingDevice,
	}
}

func (p *PodLifecycleHandlerImpl) OnNewController(ctx context.Context, controller types.Controller) error {
	spot := "provider/OnNewController"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Trace("invoking callback")
	return nil
}
func (p *PodLifecycleHandlerImpl) OnMissingController(ctx context.Context, controller types.Controller) error {
	spot := "provider/OnMissingController"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Trace("invoking callback")
	return nil
}
func (p *PodLifecycleHandlerImpl) OnExistingController(ctx context.Context, controller types.Controller) error {
	spot := "provider/OnExistingController"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Trace("invoking callback")
	return nil
}
func (p *PodLifecycleHandlerImpl) OnNewDevice(ctx context.Context, controllerName string, device types.Device) error {
	spot := "provider/OnNewDevice"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Trace("invoking callback")
	return nil
}
func (p *PodLifecycleHandlerImpl) OnMissingDevice(ctx context.Context, controllerName string, device types.Device) error {
	spot := "provider/OnMissingDevice"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Trace("invoking callback")
	return nil
}
func (p *PodLifecycleHandlerImpl) OnExistingDevice(ctx context.Context, controllerName string, device types.Device) error {
	spot := "provider/OnExistingDevice"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Trace("invoking callback")
	return nil
}
