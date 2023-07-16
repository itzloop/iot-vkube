package agent

import (
	"context"
	"github.com/itzloop/iot-vkube/internal/callback"
	"github.com/itzloop/iot-vkube/internal/pool"
	"github.com/itzloop/iot-vkube/internal/store"
	"github.com/itzloop/iot-vkube/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"time"
)

type Service struct {
	store               store.Store
	callbacks           *callback.ServiceCallBacks
	httpFetchWorkerPool *pool.WorkerPool
	diffWorkerPool      *pool.WorkerPool
	interval            time.Duration
}

func NewService(store store.Store, workerPool *pool.WorkerPool, diffWorkerPool *pool.WorkerPool, interval time.Duration) *Service {
	srv := &Service{store: store, httpFetchWorkerPool: workerPool, diffWorkerPool: diffWorkerPool, interval: interval}

	// register incoming callbacks
	srv.RegisterCallbacks(nil)

	return srv
}

func (service *Service) RegisterToCallbacks(cb callback.Callback) {
	cb.RegisterCallbacks(service.ServiceCallBacks())
}

func (service *Service) RegisterCallbacks(cb *callback.ServiceCallBacks) {
	var defaultCB = callback.DefaultServiceCallBacks()
	if cb == nil {
		service.callbacks = defaultCB
		return
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

	service.callbacks = cb
}

func (service *Service) Start(ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error { return service.agentWorker(groupCtx, service.interval) })

	go func() {
		<-groupCtx.Done()
		service.Close()
	}()

	return group.Wait()
}

func (service *Service) Close() error {
	// shutdown http server
	//err := service.server.Shutdown(context.Background())
	//if err != nil {
	//	if err != http.ErrServerClosed {
	//		return nil
	//	}
	//}

	return nil
}

// TODO
// TODO agentWorker should call following endpoints periodically
// - controller readiness
// - device readiness
func (service *Service) agentWorker(ctx context.Context, interval time.Duration) error {
	spot := "agentWorker"
	ticker := time.Tick(interval)
	entry := utils.GetEntryFromContext(ctx)
	entry = entry.WithFields(logrus.Fields{
		"spot":     spot,
		"interval": interval.String(),
	})

	ctx = utils.ContextWithEntry(ctx, entry)

	entry.Info("starting agent worker")
	defer entry.Info("exiting agent worker")

	if ticker == nil {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				entry.Info("updating state")
				if err := service.diff(ctx); err != nil {
					continue
				}
			}
		}
	}

	for {
		select {
		case <-ticker:
			entry.Info("updating state")
			if err := service.diff(ctx); err != nil {
				continue
			}
		case <-ctx.Done():
			return nil
		}
	}
}
