package provider

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	v1 "k8s.io/client-go/listers/core/v1"
	"net/http"
	"sync"
)

type PodLifecycleHandlerImpl struct {
	addr      string
	pods      sync.Map
	podLister v1.PodLister
}

func NewPodLifecycleHandlerImpl(addr string, lister v1.PodLister) *PodLifecycleHandlerImpl {
	return &PodLifecycleHandlerImpl{
		addr:      addr,
		pods:      sync.Map{},
		podLister: lister,
	}
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

	p.pods.Store(pod.Name, pod)

	return nil
}
func (p *PodLifecycleHandlerImpl) UpdatePod(ctx context.Context, pod *corev1.Pod) error {
	fmt.Println("UpdatePod", pod.Namespace, pod.Name)
	_, loaded := p.pods.LoadOrStore(pod.Name, pod)
	if loaded {
		return nil
	}

	return p.CreatePod(ctx, pod)
}
func (p *PodLifecycleHandlerImpl) DeletePod(ctx context.Context, pod *corev1.Pod) error {
	fmt.Println("DeletePod", pod.Namespace, pod.Name)

	_, loaded := p.pods.LoadAndDelete(pod.Name)
	if !loaded {
		return errors.New("pod with this name doesn't exist")
	}
	return nil
}
func (p *PodLifecycleHandlerImpl) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	fmt.Println("GetPod", namespace, name)
	return &corev1.Pod{}, nil
}

func (p *PodLifecycleHandlerImpl) GetPodStatus(ctx context.Context, namespace, name string) (*corev1.PodStatus, error) {
	fmt.Println("GetPodStatus", namespace, name)

	v, ok := p.pods.Load(name)
	if !ok {
		return nil, errors.New("pod not found")
	}

	pod, ok := v.(*corev1.Pod)
	if !ok {
		return nil, errors.New("this shouldn't happen")
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

	return &pod.Status, nil
}
func (p *PodLifecycleHandlerImpl) GetPods(context.Context) ([]*corev1.Pod, error) {
	fmt.Println("GetPods")

	var pods []*corev1.Pod
	p.pods.Range(func(key, value any) bool {
		pods = append(pods, value.(*corev1.Pod))
		return true
	})

	return pods, nil
}
