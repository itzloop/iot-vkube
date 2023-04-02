package main

import (
	"flag"
	"github.com/itzloop/iot-vkube/internal/provider"
	"github.com/pkg/errors"
	"github.com/virtual-kubelet/virtual-kubelet/node"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

func main() {
	// TODO get kubernetes client
	kubeConfigPath := flag.String("kubeconfig", "/home/loop/.kube/config", "kubernetes cluster config")
	flag.Parse()

	client, err := newKubernetesClient(*kubeConfigPath)
	if err != nil {
		panic(err)
	}

	// TODO setup pod controller
	p := provider.NewPodLifecycleHandlerImpl()
	pc, err := node.NewPodController(node.PodControllerConfig{
		PodClient:                                client.CoreV1(),
		PodInformer:                              nil,
		EventRecorder:                            nil,
		Provider:                                 p,
		ConfigMapInformer:                        nil,
		SecretInformer:                           nil,
		ServiceInformer:                          nil,
		SyncPodsFromKubernetesRateLimiter:        nil,
		SyncPodsFromKubernetesShouldRetryFunc:    nil,
		DeletePodsFromKubernetesRateLimiter:      nil,
		DeletePodsFromKubernetesShouldRetryFunc:  nil,
		SyncPodStatusFromProviderRateLimiter:     nil,
		SyncPodStatusFromProviderShouldRetryFunc: nil,
		PodEventFilterFunc:                       nil,
	})
	if err != nil {
		panic(err)
	}

	// TODO setup node controller
	// TODO setup HTTP endpoints
}

func newKubernetesClient(configPath string) (*kubernetes.Clientset, error) {
	var config *rest.Config

	// Check if the kubeConfig file exists.
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		// Get the kubeconfig from the filepath.
		config, err = clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			return nil, errors.Wrap(err, "error building client config")
		}
	} else {
		// Set to in-cluster config.
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, errors.Wrap(err, "error building in cluster config")
		}
	}

	if masterURI := os.Getenv("MASTER_URI"); masterURI != "" {
		config.Host = masterURI
	}

	return kubernetes.NewForConfig(config)
}
