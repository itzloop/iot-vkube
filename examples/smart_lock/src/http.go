package smart_lock

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/itzloop/iot-vkube/examples/smart_lock/src/routers"
	store2 "github.com/itzloop/iot-vkube/internal/store"
	"io"
	"k8s.io/apimachinery/pkg/util/json"
	"log"
	"net/http"
	"sync"
)

const (
	deviceNameKey     = "device_name"
	controllerNameKey = "controller_name"
)

type server struct {
	lcs map[string]*LockController
	mu  sync.Mutex
}

func handleError(w http.ResponseWriter, status int, msg string) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": msg,
	})
}

func (s *server) getController(w http.ResponseWriter, r *http.Request) *LockController {
	controllerName, ok := mux.Vars(r)[controllerNameKey]
	if !ok {
		handleError(w, http.StatusBadRequest, "controller name must be specified")
		return nil
	}

	c, ok := s.lcs[controllerName]
	if !ok {
		handleError(w, http.StatusNotFound, "controller not found")
		return nil
	}

	return c
}

func (s *server) getDevice(w http.ResponseWriter, r *http.Request) *SmartLock {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := s.getController(w, r)
	if c == nil {
		return nil
	}

	deviceName, ok := mux.Vars(r)[deviceNameKey]
	if !ok {
		handleError(w, http.StatusBadRequest, "device name must be specified")
		return nil
	}

	l, err := c.GetLock(deviceName)
	if err != nil {
		handleError(w, http.StatusNotFound, "device not found")
		return nil
	}

	return l
}

func (s *server) add(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var body map[string]interface{}

	c := s.getController(w, r)
	if c == nil {
		return
	}

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

	_, err = c.CreateLock(deviceName.(string))
	if err != nil {
		handleError(w, http.StatusConflict, "device already exists")
		return
	}

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte{})
}

func (s *server) get(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	l := s.getDevice(w, r)
	if l == nil {
		return
	}

	st := "locked"
	locked, _ := l.Locked()
	if !locked {
		st = "unlocked"
	}

	readiness, err := l.Readiness()
	if err != nil {
		handleError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get readiness of device: %v", readiness))
		return
	}

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":      l.Name(),
		"get":       st,
		"readiness": readiness,
	})
}

func (s *server) update(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	l := s.getDevice(w, r)
	if l == nil {
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		handleError(w, http.StatusBadRequest, fmt.Sprintf("failed to read body: %v", err))
		return
	}

	body := struct {
		Lock bool `json:"lock,omitempty"`
	}{}

	if err = json.Unmarshal(bodyBytes, &body); err != nil {
		handleError(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal body: %v", err))
		return
	}

	if body.Lock {
		if err = l.Lock(); err != nil {
			handleError(w, http.StatusBadRequest, fmt.Sprintf("failed to lock: %v", err))
			return
		}
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "locked",
		})
	} else {
		if err = l.UnLock(); err != nil {
			handleError(w, http.StatusBadRequest, fmt.Sprintf("failed to unlock: %v", err))
			return
		}
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "unlocked",
		})
	}
}

func (s *server) lock(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	l := s.getDevice(w, r)
	if l == nil {
		return
	}

	if err := l.Lock(); err != nil {
		handleError(w, http.StatusInternalServerError, fmt.Sprintf("failed to lock: %v", err))
		return
	}

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"get": "locked",
	})
}

func (s *server) unlock(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l := s.getDevice(w, r)
	if l == nil {
		return
	}

	if err := l.UnLock(); err != nil {
		handleError(w, http.StatusInternalServerError, fmt.Sprintf("failed to unlock: %v", err))
		return
	}

	w.Header().Add("Access-Control-Allow-Origin", "*")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"get": "unlocked",
	})
}

func (s *server) list(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := s.getController(w, r)
	if c == nil {
		return
	}

	resp := struct {
		Name      string `json:"name,omitempty"`
		Readiness bool   `json:"readiness,omitempty"`
		Devices   []struct {
			Name      string `json:"name,omitempty"`
			Readiness bool   `json:"readiness,omitempty"`
		} `json:"devices"`
	}{
		Name:      c.name,
		Readiness: c.readiness,
		Devices: []struct {
			Name      string `json:"name,omitempty"`
			Readiness bool   `json:"readiness,omitempty"`
		}{},
	}

	devices, err := c.ListLocks()
	if err != nil {
		handleError(w, http.StatusInternalServerError, fmt.Sprintf("failed to list locks: %v", err))
		return
	}

	for _, device := range devices {
		readiness, err := device.Readiness()
		if err != nil {
			handleError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get readiness of lock: %v", err))
			return
		}
		resp.Devices = append(resp.Devices, struct {
			Name      string `json:"name,omitempty"`
			Readiness bool   `json:"readiness,omitempty"`
		}{Name: device.Name(), Readiness: readiness})
	}

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *server) listControllers(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var resp = []struct {
		Name      string `json:"name,omitempty"`
		Readiness bool   `json:"readiness,omitempty"`
	}{}

	for _, lc := range s.lcs {
		resp = append(resp, struct {
			Name      string `json:"name,omitempty"`
			Readiness bool   `json:"readiness,omitempty"`
		}{Name: lc.name, Readiness: lc.GetReadiness()})
	}

	w.Header().Add("Access-Control-Allow-Origin", "*")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *server) toggleControllerReadiness(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := s.getController(w, r)
	if c == nil {
		return
	}

	c.SetReadiness(!c.GetReadiness())

	w.Header().Add("Access-Control-Allow-Origin", "*")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

func (s *server) addController(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

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

	nameInterface, ok := body["name"]
	if !ok {
		handleError(w, http.StatusBadRequest, "name not found in body")
		return
	}

	name, ok := nameInterface.(string)
	if !ok {
		handleError(w, http.StatusBadRequest, "name must be of type string")
		return
	}

	readinessInterface, ok := body["readiness"]
	if !ok {
		handleError(w, http.StatusBadRequest, "readiness not found in body")
		return
	}

	readiness, ok := readinessInterface.(bool)
	if !ok {
		handleError(w, http.StatusBadRequest, "readiness must be of type bool")
		return
	}

	_, ok = s.lcs[name]
	if ok {
		handleError(w, http.StatusConflict, "controller already exists")
		return
	}

	s.lcs[name] = NewLockController(name, readiness)
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte{})
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func RunServer(addr string) {
	fmt.Printf("server is listening on %s\n", addr)
	//srv := server{lcs: map[string]*LockController{}}

	store := store2.NewLocalStoreImpl()
	r := gin.Default()
	r.Use(CORSMiddleware())
	controllers := r.Group("/controllers")
	{
		controllersRouteHandlers := routers.NewControllersRouteHandler(store)
		controllers.GET("", controllersRouteHandlers.List)
		controllers.POST("", controllersRouteHandlers.Create)
		controller := controllers.Group("/:controller_name")
		{
			controller.GET("", controllersRouteHandlers.Get)
			controller.PATCH("", controllersRouteHandlers.Update)
			controller.DELETE("", controllersRouteHandlers.Delete)

			devicesRouterHandlers := routers.NewDevicesRouteHandler(store)
			devices := controller.Group("/devices")
			{
				devices.GET("", devicesRouterHandlers.List)
				devices.POST("", devicesRouterHandlers.Create)
				device := devices.Group("/:device_name")
				{
					device.GET("", devicesRouterHandlers.Get)
					device.PATCH("", devicesRouterHandlers.Update)
					device.DELETE("", devicesRouterHandlers.Delete)
				}
			}
		}
	}

	//r := mux.NewRouter()

	//controllerRouter := r.
	//	PathPrefix(fmt.Sprintf("/controllers/{controller_name}")).
	//	Subrouter()
	//
	//devicesRouter := controllerRouter.
	//	PathPrefix("/devices").
	//	Subrouter()
	//
	//devicesRouter.
	//	HandleFunc("", srv.add).
	//	Methods(http.MethodPost)
	//
	//devicesRouter.
	//	HandleFunc("/{device_name}", srv.get).
	//	Methods(http.MethodGet)
	//
	//devicesRouter.
	//	HandleFunc("/{device_name}", srv.update).
	//	Methods(http.MethodPatch)
	//
	//controllerRouter.HandleFunc("", srv.list).
	//	Methods(http.MethodGet)
	//
	//r.Use(loggingMiddleware)
	//
	//r.PathPrefix("/controllers").
	//	HandlerFunc(srv.listControllers).
	//	Methods(http.MethodGet, http.MethodOptions)
	//
	//r.PathPrefix("/controllers").
	//	HandlerFunc(srv.addController).
	//	Methods(http.MethodPost)

	//http.Handle("/", r)
	//log.Fatal(http.ListenAndServe(addr, nil))

	log.Fatal(r.Run(addr))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
