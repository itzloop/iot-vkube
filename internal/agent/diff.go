package agent

import (
	"context"
	"fmt"
	"github.com/itzloop/iot-vkube/internal/pool"
	"github.com/itzloop/iot-vkube/types"
	"github.com/itzloop/iot-vkube/utils"
	"sync/atomic"
)

const (
	jobsBufferCountDefault         = 10
	httpFetchJobBufferCountDefault = 10
	httpFetchBatchCountDefault     = 1000
)

type controllerDiffModel struct {
	NewControllers      map[string]types.Controller
	MissingControllers  map[string]types.Controller
	ExistingControllers map[string]types.Controller
}

type httpFetchJob struct {
	workerPool      *pool.WorkerPool
	controller      types.Controller
	service         *Service
	batchCount      int64
	jobsBufferCount int64
}

//goland:noinspection GoUnreachableCode
func (h *httpFetchJob) Run(ctx context.Context) error {
	spot := "httpFetchJob.Run"
	ctx = utils.ContextWithSpot(ctx, spot)
	entry := utils.GetEntryFromContext(ctx)

	var (
		itr    = int64(0)
		closed = atomic.Bool{}
	)

	jobsChan := make(chan pool.Job, h.jobsBufferCount)
	defer func() {
		if closed.Load() {
			return
		}

		close(jobsChan)
	}()

	errChan, err := h.service.workerPool.ExecuteBatch(jobsChan)
	if err != nil {
		entry.WithField("error", err).
			Error("failed to execute jobs in batch")
		return err
	}

	for {
		var (
			remoteDevices    = []types.Device{}
			remoteDevicesMap = map[string]types.Device{}
		)
		select {
		case <-ctx.Done():
			if ctx.Err() == context.Canceled {
				return nil
			}

			return ctx.Err()
		default:
			devicesUrl := fmt.Sprintf("http://%s/controllers/%s/devices?itr=%d&count=%d", h.controller.Host, h.controller.Name, itr, h.batchCount)
			if err := doGetRequest(devicesUrl, &remoteDevices); err != nil {
				entry.WithField("error", err).
					Debug("failed to get remote devices, change readiness of all devices connected to this controller")
				// change readiness of all devices connected to this controller
				for _, device := range h.controller.Devices {
					remoteDevicesMap[device.Name] = types.Device{
						Name:      device.Name,
						Readiness: false,
					}
				}
			}

			itr += h.batchCount

			// end
			if err == nil && len(remoteDevices) == 0 {
				goto afterLoop
			}

			jobsChan <- &diffJob{
				controller:       h.controller,
				service:          h.service,
				remoteDevices:    remoteDevices,
				remoteDevicesMap: remoteDevicesMap,
			}
		}
	}
afterLoop:
	closed.Store(true)
	close(jobsChan)

	for err := range errChan {
		entry.WithField("error", err).
			Error("failed to handle job")
	}

	return nil
}

type diffJob struct {
	controller       types.Controller
	service          *Service
	remoteDevices    []types.Device
	remoteDevicesMap map[string]types.Device
}

func (d *diffJob) Run(ctx context.Context) error {
	spot := "diffJob.Run"
	ctx = utils.ContextWithSpot(ctx, spot)
	entry := utils.GetEntryFromContext(ctx)

	var localDevicesMap = map[string]types.Device{}

	// convert remote controllers to map
	for _, device := range d.remoteDevices {
		d.remoteDevicesMap[device.Name] = device
	}

	// convert local devices to map
	for _, device := range d.controller.Devices {
		localDevicesMap[device.Name] = device
	}

	if err := d.service.processDeviceDiff(ctx, d.controller.Name, d.service.deviceDiff(ctx, localDevicesMap, d.remoteDevicesMap)); err != nil {
		entry.WithField("error", err).
			Error("failed to get process deviceDiff")
		return err
	}

	return nil
}

func (service *Service) diff(ctx context.Context) error {
	spot := "diff"
	ctx = utils.ContextWithSpot(ctx, spot)
	entry := utils.GetEntryFromContext(ctx)

	//// get remote controllers
	//var remoteControllers []ControllerBody
	//var remoteControllersMap = map[string]ControllerBody{}
	//controllersUrl := fmt.Sprintf("http://%s/controllers", hook)
	//if err := doGetRequest(controllersUrl, &remoteControllers); err != nil {
	//	entry.WithFields(logrus.Fields{
	//		"error": err,
	//		"url":   controllersUrl,
	//	}).Error("failed to get remote controllers ")
	//	return err
	//}

	//// convert remote controllers to map
	//for _, controller := range remoteControllers {
	//	remoteControllersMap[controller.Name] = controller
	//}

	// get local controllers
	//localControllers, err := service.store.GetControllersMap(ctx)
	//if err != nil {
	//	entry.WithField("error", err).
	//		Error("failed to get local controllers ")
	//	return err
	//}
	//
	//if err = service.processControllerDiff(ctx, service.controllerDiff(ctx, localControllers, remoteControllersMap)); err != nil {
	//	entry.WithField("error", err).
	//		Error("failed to get process controllerDiff")
	//	return err
	//}

	// now that controllers are synced get the controllerDiff for devices on each controller
	controllers, err := service.store.GetControllers(ctx)
	if err != nil {
		entry.WithField("error", err).
			Error("failed to get controllers ")
		return err
	}

	closed := atomic.Bool{}
	jobsChan := make(chan pool.Job, httpFetchJobBufferCountDefault)
	defer func() {
		if closed.Load() {
			return
		}

		close(jobsChan)
	}()

	errChan, err := service.workerPool.ExecuteBatch(jobsChan)
	if err != nil {
		entry.WithField("error", err).
			Error("failed to execute jobs in batch")
		return err
	}

	for _, controller := range controllers {
		jobsChan <- &httpFetchJob{
			workerPool:      service.workerPool,
			controller:      controller,
			service:         service,
			batchCount:      httpFetchBatchCountDefault,
			jobsBufferCount: jobsBufferCountDefault,
		}
	}

	closed.Swap(true)
	close(jobsChan)

	for err := range errChan {
		entry.WithField("error", err).
			Error("failed to handle job")
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
			localController.Readiness = remoteController.Readiness
			existingControllers[remoteController.Name] = localController
		} else {
			newControllers[remoteController.Name] = types.Controller{
				// TODO Host:    ,
				// TODO Meta:    nil,
				Name:      remoteController.Name,
				Readiness: remoteController.Readiness,
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
		missingController.Readiness = false
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

func (service *Service) deviceDiff(ctx context.Context, local map[string]types.Device, remote map[string]types.Device) deviceDiffModel {
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
			localDevice.Readiness = remoteDevice.Readiness
			existingDevices[remoteDevice.Name] = localDevice
		} else {
			newDevices[remoteDevice.Name] = types.Device{
				Name:      remoteDevice.Name,
				Readiness: remoteDevice.Readiness,
			}
		}
	}

	// find missing devices
	for _, localDevice := range local {
		remoteDevice, ok := remote[localDevice.Name]
		if !ok {
			missingDevices[remoteDevice.Name] = types.Device{
				Name:      localDevice.Name,
				Readiness: localDevice.Readiness,
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
			entry.WithError(err).Error("failed to call callback")
			return
		}

		// add it to local store
		if err = service.store.RegisterDevice(ctx, controllerName, newDevice); err != nil {
			entry.WithError(err).Error("failed to register device")
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
		missingDevice.Readiness = false
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
