package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/itzloop/iot-vkube/internal/callback"
	"github.com/itzloop/iot-vkube/internal/store"
	"github.com/itzloop/iot-vkube/types"
	"github.com/itzloop/iot-vkube/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/virtual-kubelet/virtual-kubelet/node"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/scheme"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"
	"sync"
)

type PodLifecycleHandlerImpl struct {
	addr      string
	podLister v1.PodLister
	selector  labels.Selector
	store     store.ReadOnlyStore

	// callbacks
	cbs *callback.ServiceCallBacks

	// event recorder
	recorder record.EventRecorder

	// node
	updateNodeMu *sync.Mutex
	node         *corev1.Node
	nnp          *node.NaiveNodeProviderV2
}

func NewPodLifecycleHandlerImpl(
	addr string,
	lister v1.PodLister,
	selector labels.Selector,
	store store.ReadOnlyStore,
	broadcaster record.EventBroadcaster,
	node *corev1.Node,
	nnp *node.NaiveNodeProviderV2) *PodLifecycleHandlerImpl {

	recorder := broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "IoT-Provider"})

	p := &PodLifecycleHandlerImpl{
		addr:         addr,
		podLister:    lister,
		selector:     selector,
		store:        store,
		recorder:     recorder,
		updateNodeMu: &sync.Mutex{},
		node:         node,
		nnp:          nnp,
	}

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

	if pod.Spec.NodeName != p.node.Name {
		entry.Trace("sync node name")
		pod.Spec.NodeName = p.node.Name
	}

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
			Host:      controllerAddress,
			Meta:      pod.Labels,
			Name:      controllerName,
			Readiness: false,
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
			Meta:      pod.Labels,
			Name:      pod.Name,
			Readiness: false,
		})
		if err != nil {
			entry.WithField("error", err).Error("failed to invoke callback OnNewDevice")
			return err
		}

		entry.Trace("updating node")
		return p.updateNode(ctx, -1)
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
		entry.WithField("error", err).Info("pod already exists, ignoring...")
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

	controllerName, ok := pod.Annotations["controllerName"]
	if !ok {
		err := errors.New("label 'controllerName' is missing")
		entry.WithField("error", err).Error("failed to read controllerName")
		return err
	}

	device, err := p.store.GetDevice(ctx, controllerName, pod.Name)
	if err != nil {
		entry.WithField("error", err).Error("device does not exist")
		return err
	}

	entry.Trace("deleting pod")
	err = p.cbs.OnDeviceDeleted(ctx, controllerName, device)
	if err != nil {
		return err
	}

	entry.Trace("updating node")
	return p.updateNode(ctx, 1)
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
	pod, err := p.podLister.Pods(namespace).Get(name)

	pJson, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}

	entry.Trace(string(pJson))
	return pod, err
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

	if pod == nil {
		entry.Trace("pod was nil without error")
		return nil, err
	}

	if pod.Spec.NodeName != p.node.Name {
		entry.Trace("sync node name")
		pod.Spec.NodeName = p.node.Name
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

	setPodPhase(pod, device)
	status := getConditionStatus(pod, device)

	entry = entry.WithFields(logrus.Fields{
		"readiness": device.Readiness,
		"status":    status,
		"phase":     pod.Status.Phase,
	})
	entry.Trace("pod status")

	changed := setPodConditions(status, pod)
	setPodContainerStatuses(pod, device)

	if !changed {
		return &pod.Status, nil
	}

	// send event if status changed
	if status == corev1.ConditionTrue {
		// send ready event
		p.recorder.Event(pod, corev1.EventTypeNormal, "ReasonReady", "successfully checked device readiness")
	} else {
		// send not ready event
		p.recorder.Event(pod, corev1.EventTypeWarning, "ReasonNotReady", "couldn't check device readiness")
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

func (p *PodLifecycleHandlerImpl) updateNode(ctx context.Context, podsDelta int64) error {
	p.updateNodeMu.Lock()
	defer p.updateNodeMu.Unlock()

	entry := utils.GetEntryFromContext(ctx)
	entry = entry.WithField("spot", "PodLifecycleHandlerImpl.updateNode")

	// copy the node first
	n := new(corev1.Node)
	p.node.DeepCopyInto(n)

	// modify the node
	pods, _ := n.Status.Allocatable.Pods().AsInt64()
	pods += podsDelta

	n.Status.Allocatable = corev1.ResourceList{
		corev1.ResourceCPU:    n.Status.Allocatable.Cpu().DeepCopy(),
		corev1.ResourceMemory: n.Status.Allocatable.Memory().DeepCopy(),
		corev1.ResourcePods:   resource.MustParse(fmt.Sprint(pods)),
	}

	// store the updated node
	p.node = n

	// update the node
	err := p.nnp.UpdateStatus(ctx, n)
	if err != nil {
		entry.WithError(err).Error("failed to update node")
		return err
	}

	return nil
}
