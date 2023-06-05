package agent

import (
	"context"
	"fmt"
	"github.com/itzloop/iot-vkube/internal/utils"
	"github.com/itzloop/iot-vkube/types"
	"github.com/sirupsen/logrus"
)

type controllerDiffModel struct {
	NewControllers      map[string]types.Controller
	MissingControllers  map[string]types.Controller
	ExistingControllers map[string]types.Controller
}

func (service *Service) diff(ctx context.Context, hook string) error {
	spot := "diff"
	ctx = utils.ContextWithSpot(ctx, spot)
	entry := utils.GetEntryFromContext(ctx)

	// get remote controllers
	var remoteControllers []ControllerBody
	var remoteControllersMap = map[string]ControllerBody{}
	controllersUrl := fmt.Sprintf("http://%s/controllers", hook)
	if err := doGetRequest(controllersUrl, &remoteControllers); err != nil {
		entry.WithFields(logrus.Fields{
			"error": err,
			"url":   controllersUrl,
		}).Error("failed to get remote controllers ")
		return err
	}

	// convert remote controllers to map
	for _, controller := range remoteControllers {
		remoteControllersMap[controller.Name] = controller
	}

	// get local controllers
	localControllers, err := service.store.GetControllersMap(ctx)
	if err != nil {
		entry.WithField("error", err).
			Error("failed to get local controllers ")
		return err
	}

	if err = service.processControllerDiff(ctx, service.controllerDiff(ctx, localControllers, remoteControllersMap)); err != nil {
		entry.WithField("error", err).
			Error("failed to get process controllerDiff")
		return err
	}

	// now that controllers are synced get the controllerDiff for devices on each controller
	controllers, err := service.store.GetControllers(ctx)
	if err != nil {
		entry.WithField("error", err).
			Error("failed to get controllers ")
		return err
	}

	for _, controller := range controllers {
		var (
			remoteDevices    DeviceListBody
			remoteDevicesMap = map[string]DeviceBody{}
			localDevicesMap  = map[string]types.Device{}
		)
		devicesUrl := fmt.Sprintf("http://%s/controllers/%s", hook, controller.Name)
		if err := doGetRequest(devicesUrl, &remoteDevices); err != nil {
			entry.WithField("error", err).
				Error("failed to get remote devices")
			continue
		}

		// convert remote controllers to map
		for _, device := range remoteDevices.Devices {
			remoteDevicesMap[device.Name] = device
		}

		// convert local devices to map
		for _, device := range controller.Devices {
			localDevicesMap[device.Name] = device
		}

		if err = service.processDeviceDiff(ctx, controller.Name, service.deviceDiff(ctx, localDevicesMap, remoteDevicesMap)); err != nil {
			entry.WithField("error", err).
				Error("failed to get process deviceDiff")
			return err
		}
	}

	return nil
}

func (service *Service) controllerDiff(ctx context.Context, local map[string]types.Controller, remote map[string]ControllerBody) controllerDiffModel {
	spot := "controllerDiff"
	ctx = utils.ContextWithSpot(ctx, spot)

	// find the controllerDiff between two lists
	var (
		existingControllers = map[string]types.Controller{}
		missingControllers  = map[string]types.Controller{}
		newControllers      = map[string]types.Controller{}
	)

	// find new and existing controllers
	for _, remoteController := range remote {
		localController, ok := local[remoteController.Name]
		if ok {
			// get readiness from remote
			// TODO maybe handle some metadata later
			localController.Ready = remoteController.Readiness
			existingControllers[remoteController.Name] = localController
		} else {
			newControllers[remoteController.Name] = types.Controller{
				// TODO Host:    ,
				// TODO Meta:    nil,
				Name:  remoteController.Name,
				Ready: remoteController.Readiness,
			}
		}
	}

	// find missing controllers
	for _, localController := range local {
		remoteController, ok := remote[localController.Name]
		if !ok {
			missingControllers[remoteController.Name] = localController
		}
	}

	return controllerDiffModel{
		NewControllers:      newControllers,
		MissingControllers:  missingControllers,
		ExistingControllers: existingControllers,
	}
}

func (service *Service) processControllerDiff(ctx context.Context, d controllerDiffModel) (err error) {
	spot := "processControllerDiff"
	ctx = utils.ContextWithSpot(ctx, spot)
	entry := utils.GetEntryFromContext(ctx)

	// update the state of the cluster:
	// for each new controller, add it to cluster
	for _, newController := range d.NewControllers {
		// TODO a callback from provider to update cluster. for now update local store
		if err = service.callbacks.OnNewController(utils.ContextWithEntry(ctx, entry.WithField("callback", "OnNewController")), newController); err != nil {
			return
		}

		// add it to local store
		if err = service.store.RegisterController(ctx, newController); err != nil {
			return
		}
	}

	// for each missing controller, change it's state to NOT-READY
	for _, missingController := range d.MissingControllers {
		// TODO a callback from provider to update cluster. for now update local store
		if err = service.callbacks.OnMissingController(utils.ContextWithEntry(ctx, entry.WithField("callback", "OnMissingController")), missingController); err != nil {
			return
		}
		// update local store
		missingController.Ready = false
		if err = service.store.UpdateController(ctx, missingController); err != nil {
			return
		}
	}

	// for each existing controller, use it's readiness state and update the cluster
	for _, existingController := range d.ExistingControllers {
		// TODO a callback from provider to update cluster. for now update local store
		if err = service.callbacks.OnExistingController(utils.ContextWithEntry(ctx, entry.WithField("callback", "OnExistingController")), existingController); err != nil {
			return
		}
		// update local store
		if err = service.store.UpdateController(ctx, existingController); err != nil {
			return
		}
	}

	return nil
}

type deviceDiffModel struct {
	NewDevices      map[string]types.Device
	MissingDevices  map[string]types.Device
	ExistingDevices map[string]types.Device
}

func (service *Service) deviceDiff(ctx context.Context, local map[string]types.Device, remote map[string]DeviceBody) deviceDiffModel {
	spot := "deviceDiff"
	ctx = utils.ContextWithSpot(ctx, spot)

	// find the deviceDiff between two lists
	var (
		existingDevices = map[string]types.Device{}
		missingDevices  = map[string]types.Device{}
		newDevices      = map[string]types.Device{}
	)

	// find new and existing devices
	for _, remoteDevice := range remote {
		localDevice, ok := local[remoteDevice.Name]
		if ok {
			// get readiness from remote
			// TODO maybe handle some metadata later
			localDevice.Ready = remoteDevice.Readiness
			existingDevices[remoteDevice.Name] = localDevice
		} else {
			newDevices[remoteDevice.Name] = localDevice
		}
	}

	// find missing devices
	for _, localDevice := range local {
		remoteDevice, ok := remote[localDevice.Name]
		if !ok {
			missingDevices[remoteDevice.Name] = types.Device{
				Name:  remoteDevice.Name,
				Ready: remoteDevice.Readiness,
			}
		}
	}

	return deviceDiffModel{
		NewDevices:      newDevices,
		MissingDevices:  missingDevices,
		ExistingDevices: existingDevices,
	}
}

func (service *Service) processDeviceDiff(ctx context.Context, controllerName string, d deviceDiffModel) (err error) {
	spot := "processDeviceDiff"
	ctx = utils.ContextWithSpot(ctx, spot)
	entry := utils.GetEntryFromContext(ctx)
	// update the state of the cluster:
	// for each new device, add it to cluster
	for _, newDevice := range d.NewDevices {
		// TODO a callback from provider to update cluster. for now update local store
		if err = service.callbacks.OnNewDevice(utils.ContextWithEntry(ctx, entry.WithField("callback", "OnNewDevice")), controllerName, newDevice); err != nil {
			return
		}

		// add it to local store
		if err = service.store.RegisterDevice(ctx, controllerName, newDevice); err != nil {
			return
		}
	}

	// for each missing device, change it's state to NOT-READY
	for _, missingDevice := range d.MissingDevices {
		// TODO a callback from provider to update cluster. for now update local store
		if err = service.callbacks.OnMissingDevice(utils.ContextWithEntry(ctx, entry.WithField("callback", "OnMissingDevice")), controllerName, missingDevice); err != nil {
			return
		}

		// update local store
		missingDevice.Ready = false
		if err = service.store.UpdateDevice(ctx, controllerName, missingDevice); err != nil {
			return
		}
	}

	// for each existing device, use it's readiness state and update the cluster
	for _, existingDevice := range d.ExistingDevices {
		// TODO a callback from provider to update cluster. for now update local store
		if err = service.callbacks.OnExistingDevice(utils.ContextWithEntry(ctx, entry.WithField("callback", "OnExistingDevice")), controllerName, existingDevice); err != nil {
			return
		}

		// update local store
		if err = service.store.UpdateDevice(ctx, controllerName, existingDevice); err != nil {
			return
		}
	}

	return nil
}
