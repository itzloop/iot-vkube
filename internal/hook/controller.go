package hook

import (
	"github.com/itzloop/iot-vkube/internal/store"
	"github.com/sirupsen/logrus"
	"net/http"
)

type ControllerService struct {
	store store.Store
}

func NewControllerService(store store.Store) *ControllerService {
	return &ControllerService{store: store}
}

func (service *ControllerService) RegisterController(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("spot", "RegisterController").Error("not implemented")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte{})
}
func (service *ControllerService) ListControllers(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("spot", "ListControllers").Error("not implemented")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte{})
}
func (service *ControllerService) GetController(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("spot", "GetController").Error("not implemented")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte{})
}
func (service *ControllerService) DeleteController(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("spot", "DeleteController").Error("not implemented")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte{})
}
func (service *ControllerService) UpdateController(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("spot", "UpdateController").Error("not implemented")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte{})
}
