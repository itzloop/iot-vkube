package agent

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/itzloop/iot-vkube/internal/store"
	"github.com/itzloop/iot-vkube/internal/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"net/http"
	"time"
)

type Service struct {
	store store.Store
	addr  string

	hooks []string

	server    *http.Server
	callbacks *ServiceCallBacks
}

func NewService(store store.Store, addr string, callbacks *ServiceCallBacks, hooks []string) *Service {
	srv := &Service{store: store, addr: addr, hooks: hooks}
	srv.RegisterCallbacks(callbacks)
	return srv
}

func (service *Service) RegisterCallbacks(cb *ServiceCallBacks) {
	var defaultCB = DefaultServiceCallBacks()
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

	service.callbacks = cb
}

// TODO
func (service *Service) Start(ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(service.httpServer)
	group.Go(func() error { return service.agentWorker(groupCtx, time.Second*5) })

	go func() {
		<-groupCtx.Done()
		service.Close()
	}()

	return group.Wait()
}

// TODO
func (service *Service) Close() error {
	// TODO handle gracefull shutdown

	// shutdown http server
	err := service.server.Shutdown(context.Background())
	if err != nil {
		if err != http.ErrServerClosed {
			return nil
		}
	}

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

	entry.Info("starting hooks worker")
	defer entry.Info("exiting hooks worker")
	for {
		select {
		case <-ticker:
			entry.Info("updating state")
			for _, hook := range service.hooks {
				if err := service.diff(ctx, hook); err != nil {
					continue
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

// TODO httpServer should handle following endpoints:
// - register controller
func (service *Service) httpServer() error {
	r := mux.NewRouter()
	service.setupControllerRoutes(r.PathPrefix("/controllers").Subrouter())
	r.Use(utils.LoggingMiddleware)

	service.server = &http.Server{
		Addr:    service.addr,
		Handler: r,
	}
	entry := logrus.WithField("addr", service.addr)
	entry.Info("server is starting...")
	defer entry.Info("exiting http server")
	return service.server.ListenAndServe()
}

func (service *Service) setupControllerRoutes(controllerRoute *mux.Router) {
	controllerService := NewControllerService(service.store)
	controllerRoute.
		Path("").
		HandlerFunc(controllerService.RegisterController).
		Methods(http.MethodPost)

	controllerRoute.
		Path("").
		HandlerFunc(controllerService.ListControllers).
		Methods(http.MethodGet)

	controllerRoute.
		Path("/{controllerName}").
		HandlerFunc(controllerService.GetController).
		Methods(http.MethodGet)

	controllerRoute.
		Path("/{controllerName}").
		HandlerFunc(controllerService.DeleteController).
		Methods(http.MethodDelete)

	controllerRoute.
		Path("/{controllerName}").
		HandlerFunc(controllerService.UpdateController).
		Methods(http.MethodPatch)
}
