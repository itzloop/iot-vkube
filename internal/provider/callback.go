package provider

import (
	"context"
	"github.com/itzloop/iot-vkube/internal/agent"
	"github.com/itzloop/iot-vkube/internal/utils"
	"github.com/itzloop/iot-vkube/types"
)

func (p *PodLifecycleHandlerImpl) AgentServiceCallBacks() *agent.ServiceCallBacks {
	return &agent.ServiceCallBacks{
		OnNewController:      p.OnNewController,
		OnMissingController:  p.OnMissingController,
		OnExistingController: p.OnExistingController,
		OnNewDevice:          p.OnNewDevice,
		OnMissingDevice:      p.OnMissingDevice,
		OnExistingDevice:     p.OnExistingDevice,
	}
}

func (p *PodLifecycleHandlerImpl) OnNewController(ctx context.Context, controller types.Controller) error {
	spot := "OnNewController"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Info("invoking callback")
	return nil
}
func (p *PodLifecycleHandlerImpl) OnMissingController(ctx context.Context, controller types.Controller) error {
	spot := "OnMissingController"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Info("invoking callback")
	return nil
}
func (p *PodLifecycleHandlerImpl) OnExistingController(ctx context.Context, controller types.Controller) error {
	spot := "OnExistingController"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Info("invoking callback")
	return nil
}
func (p *PodLifecycleHandlerImpl) OnNewDevice(ctx context.Context, device types.Device) error {
	spot := "OnNewDevice"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Info("invoking callback")
	return nil
}
func (p *PodLifecycleHandlerImpl) OnMissingDevice(ctx context.Context, device types.Device) error {
	spot := "OnMissingDevice"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Info("invoking callback")
	return nil
}
func (p *PodLifecycleHandlerImpl) OnExistingDevice(ctx context.Context, device types.Device) error {
	spot := "OnExistingDevice"
	entry := utils.GetEntryFromContext(ctx).WithField("spot", spot)
	entry.Info("invoking callback")
	return nil
}
