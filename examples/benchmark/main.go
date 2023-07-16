package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/itzloop/iot-vkube/internal/agent"
	"github.com/itzloop/iot-vkube/internal/callback"
	"github.com/itzloop/iot-vkube/internal/pool"
	"github.com/itzloop/iot-vkube/internal/store"
	"github.com/itzloop/iot-vkube/types"
	"github.com/itzloop/iot-vkube/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/json"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	var (
		d                    time.Duration
		numCpu               int
		workerPoolBuffer     int
		controllersCount     int
		devicesPerController int
		controllersBaseAddr  string
		ctx                  context.Context
		cancel               context.CancelFunc
		group                *errgroup.Group
		err                  error
		start                time.Time
		logLevel             string
	)

	flag.DurationVar(&d, "d", time.Second*30, "Test duration")
	flag.IntVar(&numCpu, "c", runtime.NumCPU(), "Core count")
	flag.IntVar(&workerPoolBuffer, "workers", 4*runtime.NumCPU(), "Worker pool count")
	flag.IntVar(&controllersCount, "controllers", 1, "Controllers count")
	flag.StringVar(&controllersBaseAddr, "caddr", "localhost:5000", "Controllers base address")
	flag.IntVar(&devicesPerController, "devices", 10, "Devices per controller")
	flag.StringVar(&logLevel, "log-level", logrus.ErrorLevel.String(), "Log level")
	flag.Parse()

	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Panicln(err)
	}

	logrus.SetLevel(lvl)

	localStore := store.NewLocalStoreImpl()
	wp := pool.NewWorkerPool(numCpu, workerPoolBuffer)
	diffWP := pool.NewWorkerPool(numCpu, workerPoolBuffer)
	srv := agent.NewService(localStore, wp, diffWP, -1)
	cbs := callbacks{
		mu:             sync.Mutex{},
		db:             map[string]struct{}{},
		t:              time.Time{},
		newDeviceCount: controllersCount * devicesPerController,
		tSet:           atomic.Bool{},
	}
	srv.RegisterCallbacks(&callback.ServiceCallBacks{
		OnNewController:      nil,
		OnMissingController:  nil,
		OnExistingController: nil,
		OnNewDevice:          cbs.onNewDevice,
		OnMissingDevice:      nil,
		OnExistingDevice:     nil,
		OnDeviceDeleted:      nil,
	})

	// generate controllers and devices
	for i := 0; i < controllersCount; i++ {
		c := types.Controller{
			Host:      controllersBaseAddr,
			Name:      fmt.Sprintf("controller-%d", i),
			Readiness: true,
			Devices:   make([]types.Device, 0, devicesPerController),
		}
		//for i := 0; i < devicesPerController; i++ {
		//	c.Devices = append(c.Devices, types.Device{
		//		Name:      fmt.Sprintf("device-%d", i),
		//		Readiness: true,
		//	})
		//}

		err := localStore.RegisterController(ctx, c)
		if err != nil {
			logrus.Panic(err)
		}
	}

	ctx, cancel = context.WithCancel(context.Background())
	group, ctx = errgroup.WithContext(ctx)

	// signal handling
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	go func(sig <-chan os.Signal, d time.Duration) {
		timer := time.NewTimer(d)
		defer timer.Stop()

		for {
			select {
			case s := <-sig:
				logrus.WithField("signal", s.String()).Info("received interrupt, quitting gracefully")
				cancel()

				fmt.Printf("Settings:\nControllers: %d\tDevicesPerController: %d\tDuration: %s\n", controllersCount, devicesPerController, d.String())
				fmt.Printf("It took %s to check for %d devices\n", cbs.t.Sub(start), controllersCount*devicesPerController)
				fmt.Println(cbs.tSet.Load())
				fmt.Println(cbs.count)
				s = <-sig
				logrus.WithField("signal", s.String()).Info("force quit")
				os.Exit(0)
			case <-timer.C:
				// done
				cancel()
				logrus.WithFields(logrus.Fields{
					"spot":     "signal",
					"duration": d.String(),
				}).Info("")
			}

		}
	}(sig, d)

	// start worker pool
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

	time.Sleep(time.Second)

	// start service
	group.Go(func() error {
		err := srv.Start(ctx)

		if err != nil {
			logrus.Info("service finished with err", err)
		} else {
			logrus.Info("service finished")
		}

		return err
	})

	start = time.Now()

	if err = group.Wait(); err != nil {
		logrus.WithField("error", err).Error("one of goroutines has been stopped")
		cancel()
		utils.WaitWithThreeDots("cleaning up", time.Second*2)
	}

	// calculate stats
	//fmt.Println(wp.GetStats())
	v := map[string]interface{}{}
	v["settings"] = struct {
		Controllers          int
		DevicesPerController int
		Duration             string
	}{
		controllersCount, devicesPerController, d.String(),
	}

	v["results"] = struct {
		Time             string
		TotalControllers int
		TotalDevices     int
	}{
		cbs.t.Sub(start).String(), controllersCount, controllersCount * devicesPerController,
	}

	js, err := json.Marshal(v)
	if err != nil {
		logrus.Panic(err)
	}

	fmt.Println(string(js))

	//fmt.Printf("Settings:\nControllers: %d\tDevicesPerController: %d\tDuration: %s\n", controllersCount, devicesPerController, d.String())
	//fmt.Printf("It took %s to check for %d devices in \n", cbs.t.Sub(start), controllersCount*devicesPerController)
}

type callbacks struct {
	mu             sync.Mutex
	db             map[string]struct{}
	t              time.Time
	tSet           atomic.Bool
	newDeviceCount int
	count          int64
}

func (cb *callbacks) onNewDevice(ctx context.Context, controllerName string, device types.Device) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.db[fmt.Sprintf("%s-%s", controllerName, device.Name)] = struct{}{}
	atomic.AddInt64(&cb.count, 1)
	if len(cb.db) >= cb.newDeviceCount {
		if cb.tSet.Swap(true) {
			return nil
		}

		cb.t = time.Now()
	}

	return nil
}
