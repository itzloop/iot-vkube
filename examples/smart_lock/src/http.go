package smart_lock

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"k8s.io/apimachinery/pkg/util/json"
	"log"
	"net/http"
	"sync"
)

const (
	deviceNameKey = "device_name"
)

type server struct {
	lc *LockController
	mu sync.Mutex
}

func handleError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": msg,
	})
}

func (s *server) getDevice(w http.ResponseWriter, r *http.Request) *SmartLock {
	deviceName, ok := mux.Vars(r)[deviceNameKey]
	if !ok {
		handleError(w, http.StatusBadRequest, "device name must be specified")
		return nil
	}

	l, err := s.lc.GetLock(deviceName)
	if err != nil {
		handleError(w, http.StatusNotFound, "device not found")
		return nil
	}

	return l
}

func (s *server) add(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}

	bodyRaw, err := io.ReadAll(r.Body)
	if err != nil {
		handleError(w, http.StatusBadRequest, fmt.Sprintf("failed to read body: %v", err))
		return
	}

	if err = json.Unmarshal(bodyRaw, &body); err != nil {
		handleError(w, http.StatusBadRequest, fmt.Sprintf("failed to read body: %v", err))
		return
	}

	deviceName, ok := body["deviceName"]
	if !ok {
		handleError(w, http.StatusBadRequest, "deviceName not found in body")
		return
	}

	_, err = s.lc.CreateLock(deviceName.(string))
	if err != nil {
		handleError(w, http.StatusNotFound, "device not found")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "created",
	})
}

func (s *server) status(w http.ResponseWriter, r *http.Request) {
	l := s.getDevice(w, r)
	if l == nil {
		return
	}

	st := "locked"
	locked, _ := l.Locked()
	if !locked {
		st = "unlocked"
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": st,
	})
}

func (s *server) lock(w http.ResponseWriter, r *http.Request) {
	l := s.getDevice(w, r)
	if l == nil {
		return
	}

	if err := l.Lock(); err != nil {
		handleError(w, http.StatusInternalServerError, fmt.Sprintf("failed to lock: %v", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "locked",
	})
}

func (s *server) unlock(w http.ResponseWriter, r *http.Request) {
	l := s.getDevice(w, r)
	if l == nil {
		return
	}

	if err := l.UnLock(); err != nil {
		handleError(w, http.StatusInternalServerError, fmt.Sprintf("failed to unlock: %v", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "unlocked",
	})
}

func (s *server) readiness(w http.ResponseWriter, r *http.Request) {
	l := s.getDevice(w, r)
	if l == nil {
		return
	}

	readiness, _ := l.Readiness()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"readiness": readiness,
	})
}
func RunServer(addr, controllerName string) {
	fmt.Printf("server is listening on %s\n", addr)
	srv := server{lc: NewLockController(controllerName)}
	funcs := map[string]func(http.ResponseWriter, *http.Request){
		"status":    srv.status,
		"lock":      srv.lock,
		"unlock":    srv.unlock,
		"readiness": srv.readiness,
	}
	methods := map[string]string{
		"status":    http.MethodGet,
		"lock":      http.MethodPatch,
		"unlock":    http.MethodPatch,
		"readiness": http.MethodGet,
	}

	r := mux.NewRouter()
	for name, f := range funcs {
		p := fmt.Sprintf("/%s/{%s}/%s", srv.lc.name, deviceNameKey, name)
		r.HandleFunc(p, f).Methods(methods[name])
	}

	r.HandleFunc(fmt.Sprintf("/%s", srv.lc.name), srv.add).
		Methods(http.MethodPost)
	r.Use(loggingMiddleware)

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
