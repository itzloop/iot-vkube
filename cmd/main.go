package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/itzloop/iot-vkube/internal/agent"
	"github.com/itzloop/iot-vkube/internal/pool"
	"github.com/itzloop/iot-vkube/internal/provider"
	"github.com/itzloop/iot-vkube/internal/stats"
	"github.com/itzloop/iot-vkube/internal/store"
	"github.com/itzloop/iot-vkube/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/virtual-kubelet/virtual-kubelet/log"
	"github.com/virtual-kubelet/virtual-kubelet/node"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func main() {
	var (
		ctx, cancel    = context.WithCancel(context.Background())
		kubeConfigPath string
		ns             string
		logLevel       string
	)

	group, ctx := errgroup.WithContext(ctx)

	flag.StringVar(&kubeConfigPath, "kubeconfig", "/home/loop/.kube/config", "kubernetes cluster config")
	flag.StringVar(&ns, "namespace", "default", "kubernetes namespace")
	flag.StringVar(&ns, "n", "default", "kubernetes namespace")
	flag.StringVar(&logLevel, "log-level", logrus.InfoLevel.String(), "log level")
	flag.Parse()

	// set log level
	if lvl, err := logrus.ParseLevel(logLevel); err != nil {
		logrus.Fatalln(err)
	} else {
		logrus.SetLevel(lvl)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	go func(sig <-chan os.Signal) {
		s := <-sig
		logrus.WithField("signal", s.String()).Info("received interrupt, quitting gracefully")
		cancel()

		s = <-sig
		logrus.WithField("signal", s.String()).Info("force quit")
		os.Exit(0)
	}(sig)

	client, err := newKubernetesClient(kubeConfigPath)
	if err != nil {
		panic(err)
	}

	//res, err := client.CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})
	//if err != nil {
	//	panic(err)
	//}
	//
	//for _, namespace := range res.Items {
	//	logrus.Info(namespace.Name)
	//}

	// get node spec
	n, err := getNodeSpec("vkube", "test-v0.0.1")
	if err != nil {
		panic(err)
	}

	// create informer
	informer := informers.NewSharedInformerFactoryWithOptions(client, time.Second*15, informers.WithNamespace(ns))

	// create event recorded
	eb := record.NewBroadcaster()
	eb.StartLogging(log.GetLogger(ctx).Infof)
	eb.StartRecordingToSink(&corev1client.EventSinkImpl{Interface: client.CoreV1().Events(ns)})

	// setup provider
	requirement, err := labels.NewRequirement("itzloop.dev/virtual-kubelet", selection.Exists, []string{})
	if err != nil {
		panic(err)
	}
	selector := labels.NewSelector().Add(*requirement)

	// setup gin
	ginEngine, startGinEngineFunc := setupGin("localhost:5001")

	// create stats provider	TODO
	_ = stats.NewStatsHandler(informer.Core().V1().Pods().Lister(), informer.Core().V1().Nodes().Lister(), selector, ginEngine)

	// create worker pool
	wp := pool.NewWorkerPool(runtime.NumCPU(), 4*runtime.NumCPU())
	diffWP := pool.NewWorkerPool(runtime.NumCPU(), 4*runtime.NumCPU())

	// create provider
	st := store.NewLocalStoreImpl()
	service := agent.NewService(st, wp, diffWP, 10*time.Second)

	// setup native node provider
	nativeProvider := node.NewNaiveNodeProvider()
	p := provider.NewPodLifecycleHandlerImpl("localhost:5000", informer.Core().V1().Pods().Lister(), selector, st, eb, n, nativeProvider)

	// register callbacks
	service.RegisterToCallbacks(p)
	p.RegisterToCallbacks(service)

	// setup pod controller
	pc, err := node.NewPodController(node.PodControllerConfig{
		PodClient:         client.CoreV1(),
		EventRecorder:     eb.NewRecorder(scheme.Scheme, corev1.EventSource{Component: path.Join(n.Name, "pod-controller")}),
		PodInformer:       informer.Core().V1().Pods(),
		Provider:          p,
		ConfigMapInformer: informer.Core().V1().ConfigMaps(),
		SecretInformer:    informer.Core().V1().Secrets(),
		ServiceInformer:   informer.Core().V1().Services(),
	})
	if err != nil {
		panic(err)
	}

	// setup node controller
	nc, err := node.NewNodeController(nativeProvider, n, client.CoreV1().Nodes())
	if err != nil {
		panic(err)
	}

	// start informer
	group.Go(func() error {
		informer.Start(ctx.Done())
		return nil
	})

	// start podController
	group.Go(func() error {
		err := pc.Run(ctx, runtime.NumCPU())
		if err != nil {
			logrus.Info("pod controller finished with err", err)
		} else {
			logrus.Info("pod controller finished")
		}
		return err
	})

	// start node controller
	group.Go(func() error {
		err := nc.Run(ctx)
		if err != nil {
			logrus.Info("node controller finished with err", err)
		} else {
			logrus.Info("node controller finished")
		}
		return err
	})

	// start http server
	group.Go(func() error {
		err := startGinEngineFunc(ctx)
		if err != nil {
			logrus.Info("gin finished with err", err)
		} else {
			logrus.Info("gin finished")
		}
		return err
	})

	group.Go(func() error {
		wp.Start(ctx)
		<-ctx.Done()
		err := wp.Close()
		if err != nil {
			logrus.Info("worker pool finished with err", err)
		} else {
			logrus.Info("worker pool finished")
		}
		return err
	})

	group.Go(func() error {
		diffWP.Start(ctx)
		<-ctx.Done()
		err := diffWP.Close()
		if err != nil {
			logrus.Info("worker pool finished with err", err)
		} else {
			logrus.Info("worker pool finished")
		}
		return err
	})

	group.Go(func() error {
		err := service.Start(ctx)

		if err != nil {
			logrus.Info("service finished with err", err)
		} else {
			logrus.Info("service finished")
		}

		return err
	})

	logrus.Info("setup complete")

	if err := group.Wait(); err != nil {
		logrus.WithField("error", err).Error("one of goroutines has been stopped")
		cancel()
		utils.WaitWithThreeDots("cleaning up", time.Second*2)
	}
}

func setupGin(addr string) (*gin.Engine, func(ctx context.Context) error) {
	r := gin.Default()
	r.Use(utils.CORSMiddleware())

	startHTTPServer := func(ctx context.Context) error {
		srv := &http.Server{
			Addr:    addr,
			Handler: r,
		}
		entry := utils.GetEntryFromContext(ctx)

		go func() {
			<-ctx.Done()
			srv.Shutdown(context.Background())
		}()

		entry.WithField("addr", addr).Info("starting http server")
		err := srv.ListenAndServe()
		if err != nil {
			if err == http.ErrServerClosed {
				entry.WithField("error", err).Info("http server closed")
				return nil
			}
		}

		return nil
	}

	return r, startHTTPServer
}

func getNodeSpec(name, version string) (*corev1.Node, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	maxPods := 10000

	return &corev1.Node{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"type":                   "virtual-kubelet",
				"kubernetes.io/role":     "agent",
				"beta.kubernetes.io/os":  strings.ToLower(runtime.GOOS),
				"kubernetes.io/hostname": hostname,
				"alpha.service-controller.kubernetes.io/exclude-balancer": "true",
			},
		},
		Spec: corev1.NodeSpec{
			Taints: nodeTaints(),
		},
		Status: corev1.NodeStatus{
			NodeInfo: corev1.NodeSystemInfo{
				OperatingSystem: runtime.GOOS,
				Architecture:    runtime.GOARCH,
				KubeletVersion:  version,
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(fmt.Sprint(runtime.NumCPU())),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
				corev1.ResourcePods:   resource.MustParse(fmt.Sprint(maxPods)),
			},
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(fmt.Sprint(runtime.NumCPU())),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
				corev1.ResourcePods:   resource.MustParse(fmt.Sprint(maxPods)),
			},
			Conditions:      nodeConditions(),
			Addresses:       nil,
			DaemonEndpoints: corev1.NodeDaemonEndpoints{},
		},
	}, nil
}

// TODO wtf?
func nodeConditions() []corev1.NodeCondition {
	return []corev1.NodeCondition{
		{
			Type:               "OutOfDisk",
			Status:             corev1.ConditionFalse,
			LastHeartbeatTime:  v1.Now(),
			LastTransitionTime: v1.Now(),
			Reason:             "KubeletHasSufficientDisk",
			Message:            "kubelet has sufficient disk space available",
		},
		{
			Type:               "MemoryPressure",
			Status:             corev1.ConditionFalse,
			LastHeartbeatTime:  v1.Now(),
			LastTransitionTime: v1.Now(),
			Reason:             "KubeletHasSufficientMemory",
			Message:            "kubelet has sufficient memory available",
		},
		{
			Type:               "DiskPressure",
			Status:             corev1.ConditionFalse,
			LastHeartbeatTime:  v1.Now(),
			LastTransitionTime: v1.Now(),
			Reason:             "KubeletHasNoDiskPressure",
			Message:            "kubelet has no disk pressure",
		},
		{
			Type:               "NetworkUnavailable",
			Status:             corev1.ConditionFalse,
			LastHeartbeatTime:  v1.Now(),
			LastTransitionTime: v1.Now(),
			Reason:             "RouteCreated",
			Message:            "RouteController created a route",
		},
		{
			Type:               "PIDPressure",
			Status:             corev1.ConditionFalse,
			LastHeartbeatTime:  v1.Now(),
			LastTransitionTime: v1.Now(),
			Reason:             "NodeHasSufficientPID",
			Message:            "NodeHasSufficientPID",
		},
		{
			Type:               "Ready",
			Status:             corev1.ConditionTrue,
			LastHeartbeatTime:  v1.Now(),
			LastTransitionTime: v1.Now(),
			Reason:             "KubeletReady",
			Message:            "kubelet is ready.",
		},
	}
}

func nodeTaints() []corev1.Taint {
	var (
		now    = v1.Now()
		effect = corev1.TaintEffectNoSchedule
		key    = "itzloop.dev/virtual-kubelet"
		value  = "true"
	)
	return []corev1.Taint{
		{
			Key:       key,
			Value:     value,
			Effect:    effect,
			TimeAdded: &now,
		},
	}
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
