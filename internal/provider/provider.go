package provider

import (
	"bytes"
	"context"
	"fmt"
	"github.com/itzloop/iot-vkube/internal/agent"
	"github.com/pkg/errors"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/json"
	v1 "k8s.io/client-go/listers/core/v1"
	"net/http"
)

type PodLifecycleHandlerImpl struct {
	addr      string
	podLister v1.PodLister
	selector  labels.Selector
	service   *agent.Service
}

func NewPodLifecycleHandlerImpl(addr string, lister v1.PodLister, selector labels.Selector, service *agent.Service) *PodLifecycleHandlerImpl {
	return &PodLifecycleHandlerImpl{
		addr:      addr,
		podLister: lister,
		selector:  selector,
		service:   service,
	}

	// TODO start the service
	// TODO register service
	// Close the service
}

func (p *PodLifecycleHandlerImpl) CreatePod(ctx context.Context, pod *corev1.Pod) error {
	fmt.Println("CreatePod", pod.Namespace, pod.Name)
	controllerName, ok := pod.Labels["controllerName"]
	if !ok {
		return errors.New("label 'controllerName' is missing")
	}

	url := fmt.Sprintf("http://%s/%s", p.addr, controllerName)
	body, err := json.Marshal(map[string]interface{}{
		"deviceName": pod.Name,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
		bodyStr, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		fmt.Printf("failed to create pod: %v\n", string(bodyStr))
		return fmt.Errorf("failed to create pod: %v", string(bodyStr))
	}

	return nil
}
func (p *PodLifecycleHandlerImpl) UpdatePod(ctx context.Context, pod *corev1.Pod) error {
	fmt.Println("UpdatePod", pod.Namespace, pod.Name)
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
	fmt.Println("DeletePod", pod.Namespace, pod.Name)
	return nil
}

// GetPod just uses the pod lister to get the pod from kuber
func (p *PodLifecycleHandlerImpl) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	fmt.Println("GetPod", namespace, name)
	return p.podLister.Pods(namespace).Get(name)
}

func (p *PodLifecycleHandlerImpl) GetPodStatus(ctx context.Context, namespace, name string) (*corev1.PodStatus, error) {
	fmt.Println("GetPodStatus", namespace, name)

	pod, err := p.GetPod(ctx, namespace, name)
	if err != nil {
		// TODO logrus
		return nil, err
	}

	controllerName, ok := pod.Labels["controllerName"]
	if !ok {
		return nil, errors.New("label 'controllerName' is missing")
	}

	url := fmt.Sprintf("http://%s/%s/%s/readiness", p.addr, controllerName, pod.Name)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		bodyStr, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("failed to create pod: %v", string(bodyStr))
	}

	bodyRaw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var body map[string]interface{}
	if err = json.Unmarshal(bodyRaw, &body); err != nil {
		return nil, err
	}

	readinessInterface, ok := body["readiness"]
	if !ok {
		return nil, errors.New("readiness not found")
	}

	readiness, ok := readinessInterface.(bool)
	if !ok {
		return nil, errors.New("readiness must be bool")
	}

	var (
		status corev1.ConditionStatus
		phase  corev1.PodPhase
	)
	if readiness {
		status = corev1.ConditionTrue
		phase = corev1.PodRunning
	} else {
		status = corev1.ConditionFalse
		phase = corev1.PodPending
	}

	fmt.Printf("status: %s phase: %s\n", status, phase)

	pod.Status.Message = string(bodyRaw)
	pod.Status.Phase = phase
	pod.Status.Conditions = append(pod.Status.Conditions, corev1.PodCondition{
		Type:   corev1.PodReady,
		Status: status,
	})

	started := true
	pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses, corev1.ContainerStatus{
		Name: "",
		State: corev1.ContainerState{
			Running: &corev1.ContainerStateRunning{
				StartedAt: metav1.Now(),
			},
		},
		Ready:        true,
		RestartCount: 0,
		Image:        pod.Spec.Containers[0].Image,
		Started:      &started,
	})

	return &pod.Status, nil
}
func (p *PodLifecycleHandlerImpl) GetPods(context.Context) ([]*corev1.Pod, error) {
	fmt.Println("GetPods")
	return p.podLister.List(p.selector)
}
