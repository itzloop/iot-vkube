package provider

import (
	"context"
	"errors"
	corev1 "k8s.io/api/core/v1"
)

type PodLifecycleHandlerImpl struct {
}

func NewPodLifecycleHandlerImpl() *PodLifecycleHandlerImpl {
	return &PodLifecycleHandlerImpl{}
}

func (p *PodLifecycleHandlerImpl) CreatePod(ctx context.Context, pod *corev1.Pod) error {
	return errors.New("not implemented")
}
func (p *PodLifecycleHandlerImpl) UpdatePod(ctx context.Context, pod *corev1.Pod) error {
	return errors.New("not implemented")
}
func (p *PodLifecycleHandlerImpl) DeletePod(ctx context.Context, pod *corev1.Pod) error {
	return errors.New("not implemented")
}
func (p *PodLifecycleHandlerImpl) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	return nil, errors.New("not implemented")
}
func (p *PodLifecycleHandlerImpl) GetPodStatus(ctx context.Context, namespace, name string) (*corev1.PodStatus, error) {

	return nil, errors.New("not implemented")
}
func (p *PodLifecycleHandlerImpl) GetPods(context.Context) ([]*corev1.Pod, error) {
	return nil, errors.New("not implemented")
}
