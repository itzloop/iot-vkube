package agent

import (
	"context"
	"fmt"
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

	server *http.Server
}

func NewService(store store.Store, addr string) *Service {
	return &Service{store: store, addr: addr}
}

// TODO
func (service *Service) Start(ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(service.httpServer)
	group.Go(func() error { return service.agentWorker(groupCtx, time.Second*15) })

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
	ticker := time.Tick(interval)
	logrus.WithField("interval", interval.String()).Info("starting hooks worker")
	defer logrus.Info("exiting hooks worker")
	for {
		select {
		case <-ticker:
			// get controllers
			// then for each controller call the registered hooks
			// TODO maybe use worker pool for calling multiple controllers at a time
			controllers, err := service.store.GetControllers(ctx)
			if err != nil {
				return err
			}

			for _, controller := range controllers {
				// check controller and device readiness
				url := fmt.Sprintf("http://%s/controllers/%s", service.addr, controller.Name)
				var res ControllerBody
				if err = doGetRequest(url, &res); err != nil {
					return err
				}

				// TODO what to do when controller is not reacy
				if !res.Readiness {
					fmt.Println("controller is not ready")
					continue
				}

				for _, dev := range res.Devices {
					// TODO what to do when device is not ready
					if !dev.Readiness {
						fmt.Println("device is not ready")
						continue
					}
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
