package routers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/itzloop/iot-vkube/internal/store"
	"github.com/itzloop/iot-vkube/types"
	"net/http"
)

type DevicesRouteHandler struct {
	store store.Store
}

func NewDevicesRouteHandler(store store.Store) *DevicesRouteHandler {
	return &DevicesRouteHandler{store: store}
}

func (router *DevicesRouteHandler) List(c *gin.Context) {
	controllerName := c.Param("controller_name")
	devices, err := router.store.GetDevices(context.Background(), controllerName)
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
	}{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if err := router.store.RegisterDevice(context.Background(), controllerName, types.Device{
		Name:      body.Name,
		Readiness: *body.Readiness,
	}); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err})
		return
	}
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
