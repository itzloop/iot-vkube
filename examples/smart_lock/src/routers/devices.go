package routers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/itzloop/iot-vkube/internal/store"
	"github.com/itzloop/iot-vkube/types"
	"net/http"
	"strconv"
	"strings"
)

type DevicesRouteHandler struct {
	store store.Store
}

func NewDevicesRouteHandler(store store.Store) *DevicesRouteHandler {
	return &DevicesRouteHandler{store: store}
}

func (router *DevicesRouteHandler) List(c *gin.Context) {
	var (
		itrStr, countStr string
		itr, count       int64
		err              error
		partialIteration bool
		devices          []types.Device
	)

	itrStr = c.Query("itr")
	countStr = c.Query("count")
	controllerName := c.Param("controller_name")

	if strings.TrimSpace(itrStr) != "" {
		partialIteration = true
		itr, err = strconv.ParseInt(itrStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
	}

	if strings.TrimSpace(countStr) != "" {
		count, err = strconv.ParseInt(countStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
	} else {
		count = 20 // TODO default
	}

	if !partialIteration {
		devices, err = router.store.GetDevices(context.Background(), controllerName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		c.JSON(http.StatusOK, devices)
		return
	}

	devices, err = router.store.GetRangeDevices(context.Background(), controllerName, itr, itr+count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, devices)
}

func (router *DevicesRouteHandler) Get(c *gin.Context) {
	controllerName := c.Param("controller_name")
	deviceName := c.Param("device_name")

	device, err := router.store.GetDevice(context.Background(), controllerName, deviceName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	c.JSON(http.StatusOK, device)
}

func (router *DevicesRouteHandler) Create(c *gin.Context) {
	controllerName := c.Param("controller_name")
	body := struct {
		Name      string `json:"name" binding:"required"`
		Readiness *bool  `json:"readiness" binding:"required"`
		Multi     bool   `json:"multi" binding:"-"`
		Max       int    `json:"max"`
	}{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !body.Multi {
		if err := router.store.RegisterDevice(context.Background(), controllerName, types.Device{
			Name:      body.Name,
			Readiness: *body.Readiness,
		}); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
	}

	for i := 0; i < body.Max; i++ {
		name := fmt.Sprintf("%s-%d", body.Name, i)
		if err := router.store.RegisterDevice(context.Background(), controllerName, types.Device{
			Name:      name,
			Readiness: *body.Readiness,
		}); err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": err})
			return
		}
	}

	c.Status(http.StatusCreated)
	c.Writer.Write([]byte{})
}

func (router *DevicesRouteHandler) Delete(c *gin.Context) {
	deviceName := c.Param("device_name")
	controllerName := c.Param("controller_name")
	if err := router.store.DeleteDevice(context.Background(), controllerName, types.Device{
		Name: deviceName,
	}); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err})
	}

	c.String(http.StatusNoContent, "")
}

func (router *DevicesRouteHandler) Update(c *gin.Context) {
	controllerName := c.Param("controller_name")
	body := struct {
		Name      string `json:"name"`
		Readiness *bool  `json:"readiness"`
	}{}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if err := router.store.UpdateDevice(context.Background(), controllerName, types.Device{
		Name:      body.Name,
		Readiness: *body.Readiness,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
}
