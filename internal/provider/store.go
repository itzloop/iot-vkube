package provider

import (
	"context"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/listers/core/v1"
	"sync"
)

type localStore struct {
	namespaceToName map[string]map[string]*corev1.Pod
	podLister       v1.PodLister
	mu              sync.RWMutex
}

func (store *localStore) existsLocallyUnsafe(namespace, name string) (bool, error) {
	localPodNames, ok := store.namespaceToName[namespace]
	if !ok {
		return false, nil
	}

	_, ok = localPodNames[name]
	return ok, nil
}

func (store *localStore) existsLocally(namespace, name string) (bool, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	return store.existsLocallyUnsafe(namespace, name)
}

func (store *localStore) Create(ctx context.Context, pod *corev1.Pod) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// make sure that the pod does not already exist in our local store
	exists, err := store.existsLocallyUnsafe(pod.Namespace, pod.Name)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("pod already exists locally")
	}

	// now that we are sure pod does not exist in our local store
	// we can add it to our store. we might need to create the map
	// so take that into account
	localPods, ok := store.namespaceToName[pod.Namespace]
	if !ok {
		store.namespaceToName[pod.Namespace] = map[string]*corev1.Pod{}
		localPods = store.namespaceToName[pod.Namespace]
	}

	localPods[pod.Name] = pod
	return nil
}

func (store *localStore) Update(ctx context.Context, pod *corev1.Pod) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// check if we have this pod in our local store
	exists, err := store.existsLocallyUnsafe(pod.Name, pod.Namespace)
	if err != nil {
		return err
	}

	// if this pod is not in our local store then add it
	if !exists {
		return store.Create(ctx, pod)
	}

	// if this pod is in our local store then just replace it
	localPods := store.namespaceToName[pod.Namespace]
	localPods[pod.Name] = pod
	return nil
}

func (store *localStore) Delete(ctx context.Context, pod *corev1.Pod) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	delete(store.namespaceToName[pod.Namespace], pod.Name)
	return nil
}

//func (store *localStore) Get(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
//	store.mu.RLock()
//	defer store.mu.RUnlock()
//
//
//}
//
//func (store *localStore) List(ctx context.Context) (*corev1.Pod, error) {
//	store.mu.RLock()
//	defer store.mu.RUnlock()
//
//	labels.NewSelector().Add()
//	store.podLister.List()
//}
