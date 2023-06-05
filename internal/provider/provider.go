package provider

import (
	"context"
	"fmt"
	"github.com/itzloop/iot-vkube/internal/callback"
	"github.com/itzloop/iot-vkube/internal/store"
	"github.com/itzloop/iot-vkube/internal/utils"
	"github.com/itzloop/iot-vkube/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	v1 "k8s.io/client-go/listers/core/v1"
)

type PodLifecycleHandlerImpl struct {
	addr      string
	podLister v1.PodLister
	selector  labels.Selector
	store     store.ReadOnlyStore

	// callbacks
	cbs *callback.ServiceCallBacks
}

func NewPodLifecycleHandlerImpl(
	addr string,
	lister v1.PodLister,
	selector labels.Selector,
	store store.ReadOnlyStore) *PodLifecycleHandlerImpl {

	p := &PodLifecycleHandlerImpl{
		addr:      addr,
		podLister: lister,
		selector:  selector,
		store:     store,
	}

	//// register to agent callbacks
	//p.agentCallback.RegisterCallbacks(p.ServiceCallBacks())

	// register incoming callbacks
	p.RegisterCallbacks(nil)

	return p
}

func (p *PodLifecycleHandlerImpl) RegisterToCallbacks(cb callback.Callback) {
	cb.RegisterCallbacks(p.ServiceCallBacks())
}

func (p *PodLifecycleHandlerImpl) RegisterCallbacks(cb *callback.ServiceCallBacks) {
	var defaultCB = callback.DefaultServiceCallBacks()
	if cb == nil {
		cb = defaultCB
	}

	if cb.OnNewController == nil {
		cb.OnNewController = defaultCB.OnNewController
	}

	if cb.OnMissingController == nil {
		cb.OnMissingController = defaultCB.OnMissingController
	}

	if cb.OnExistingController == nil {
		cb.OnExistingController = defaultCB.OnExistingController
	}

	if cb.OnNewDevice == nil {
		cb.OnNewDevice = defaultCB.OnNewDevice
	}

	if cb.OnMissingDevice == nil {
		cb.OnMissingDevice = defaultCB.OnMissingDevice
	}

	if cb.OnExistingDevice == nil {
		cb.OnExistingDevice = defaultCB.OnExistingDevice
	}

	p.cbs = cb
}

func (p *PodLifecycleHandlerImpl) CreatePod(ctx context.Context, pod *corev1.Pod) error {
	spot := "CreatePod"
	entry := logrus.WithFields(logrus.Fields{
		"spot":      spot,
		"name":      pod.Name,
		"namespace": pod.Namespace,
	})

	ctx = utils.ContextWithEntry(ctx, entry)
	entry.Trace("creating pod")

	controllerName, ok := pod.Annotations["controllerName"]
	if !ok {
		err := errors.New("label 'controllerName' is missing")
		entry.WithField("error", err).Error("failed to read controllerName")
		return err
	}

	controllerAddress, ok := pod.Annotations["controllerAddress"]
	if !ok {
		err := errors.New("label 'controllerAddress' is missing")
		entry.WithField("error", err).Error("failed to read controllerAddress")
		return err
	}

	// check if controller exist
	controller, err := p.store.GetController(ctx, controllerName)
	if err != nil {
		entry.WithField("error", err).Info("controller not found. notifying agent...")
		controller = types.Controller{
			Host:  controllerAddress,
			Meta:  pod.Labels,
			Name:  controllerName,
			Ready: false,
		}
		err = p.cbs.OnNewController(ctx, controller)
		if err != nil {
			entry.WithField("error", err).Error("failed to notify agent")
			return err
		}
	}

	_, err = p.store.GetDevice(ctx, controllerName, pod.Name)
	if err != nil {
		entry.WithField("error", err).Debug("device does not exist")

		// device does not exist
		err = p.cbs.OnNewDevice(ctx, controllerName, types.Device{
			Meta:  pod.Labels,
			Name:  pod.Name,
			Ready: false,
		})
		if err != nil {
			entry.WithField("error", err).Error("failed to invoke callback OnNewDevice")
			return err
		}

		return nil
	}

	entry.Debug("device exists")
	return nil
}
func (p *PodLifecycleHandlerImpl) UpdatePod(ctx context.Context, pod *corev1.Pod) error {
	spot := "UpdatePod"
	entry := logrus.WithFields(logrus.Fields{
		"spot":      spot,
		"name":      pod.Name,
		"namespace": pod.Namespace,
	})

	ctx = utils.ContextWithEntry(ctx, entry)
	entry.Trace("updating pod")

	pod, err := p.GetPod(ctx, pod.Namespace, pod.Name)
	if err != nil {
		fmt.Println("pod already exists, ignoring...")
		return nil
	}

	return p.CreatePod(ctx, pod)
}

// DeletePod is on-op for now since we are using kuber's state
// later we might call and endpoint to notify another program
// that this pod has been deleted so that the program can act
// upon that action
func (p *PodLifecycleHandlerImpl) DeletePod(ctx context.Context, pod *corev1.Pod) error {
	spot := "DeletePod"
	entry := logrus.WithFields(logrus.Fields{
		"spot":      spot,
		"name":      pod.Name,
		"namespace": pod.Namespace,
	})

	ctx = utils.ContextWithEntry(ctx, entry)
	entry.Trace("deleting pod")
	return nil
}

// GetPod just uses the pod lister to get the pod from kuber
func (p *PodLifecycleHandlerImpl) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	spot := "GetPod"
	entry := logrus.WithFields(logrus.Fields{
		"spot":      spot,
		"name":      name,
		"namespace": namespace,
	})

	ctx = utils.ContextWithEntry(ctx, entry)
	entry.Trace("getting pod")
	return p.podLister.Pods(namespace).Get(name)
}

func (p *PodLifecycleHandlerImpl) GetPodStatus(ctx context.Context, namespace, name string) (*corev1.PodStatus, error) {
	spot := "GetPodStatus"
	entry := logrus.WithFields(logrus.Fields{
		"spot":      spot,
		"name":      name,
		"namespace": namespace,
	})

	ctx = utils.ContextWithEntry(ctx, entry)
	entry.Trace("getting pod status")

	pod, err := p.GetPod(ctx, namespace, name)
	if err != nil {
		entry.WithField("error", err).Error("failed to get pod")
		return nil, err
	}

	controllerName, ok := pod.Annotations["controllerName"]
	if !ok {
		err = errors.New("label 'controllerName' is missing")
		entry.WithField("error", err).Error("failed to get 'controllerName'")
		return nil, err
	}

	// get device from store
	device, err := p.store.GetDevice(ctx, controllerName, pod.Name)
	if err != nil {
		entry.WithField("error", err).Error("failed to get device")
		return nil, err
	}

	var (
		status corev1.ConditionStatus
		phase  corev1.PodPhase
	)
	if device.Ready {
		status = corev1.ConditionTrue
		phase = corev1.PodRunning
	} else {
		status = corev1.ConditionFalse
		phase = corev1.PodPending
	}

	entry = entry.WithFields(logrus.Fields{
		"readiness": device.Ready,
		"status":    status,
		"phase":     phase,
	})

	entry.Trace("setting status")

	pod.Status.Message = "TODO: what goes here?"
	pod.Status.Phase = phase

	// get last condition check if it's different from what we have now
	// if yes then update the condition otherwise ignore it
	conditions := pod.Status.Conditions
	lastCondition := conditions[len(conditions)-1]
	if lastCondition.Status != status {
		pod.Status.Conditions = append(pod.Status.Conditions, corev1.PodCondition{
			Type:   corev1.PodReady,
			Status: status,
		})
	}

	pod.Status.ContainerStatuses = nil
	for _, container := range pod.Spec.Containers {
		started := true
		pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses, corev1.ContainerStatus{
			Name: container.Name,
			State: corev1.ContainerState{
				Running: &corev1.ContainerStateRunning{
					StartedAt: metav1.Now(),
				},
			},
			Ready:        true,
			RestartCount: 0,
			Image:        container.Image,
			Started:      &started,
		})
	}

	return &pod.Status, nil
}
func (p *PodLifecycleHandlerImpl) GetPods(ctx context.Context) ([]*corev1.Pod, error) {
	spot := "GetPods"
	entry := logrus.WithFields(logrus.Fields{
		"spot":     spot,
		"selector": p.selector.String(),
	})

	ctx = utils.ContextWithEntry(ctx, entry)
	entry.Trace("listing pods")
	return p.podLister.List(p.selector)
}
