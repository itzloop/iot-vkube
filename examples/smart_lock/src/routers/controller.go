package routers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/itzloop/iot-vkube/internal/store"
	"github.com/itzloop/iot-vkube/types"
	"net/http"
)

type ControllersRouteHandler struct {
	store store.Store
}

func NewControllersRouteHandler(store store.Store) *ControllersRouteHandler {
	return &ControllersRouteHandler{store: store}
}

func (router *ControllersRouteHandler) List(c *gin.Context) {
	controllers, err := router.store.GetControllers(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, controllers)
}

func (router *ControllersRouteHandler) Get(c *gin.Context) {
	controllerName := c.Param("controller_name")

	controller, err := router.store.GetController(context.Background(), controllerName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "controller not found"})
		return
	}

	c.JSON(http.StatusOK, controller)
}

func (router *ControllersRouteHandler) Create(c *gin.Context) {
	body := struct {
		Name      string `json:"name" binding:"required"`
		Host      string `json:"host" binding:"required"`
		Readiness *bool  `json:"readiness" binding:"required"`
	}{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if err := router.store.RegisterController(context.Background(), types.Controller{
		Host:      body.Host,
		Name:      body.Name,
		Readiness: *body.Readiness,
	}); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err})
		return
	}
}

func (router *ControllersRouteHandler) Delete(c *gin.Context) {
	controllerName := c.Param("controller_name")
	if err := router.store.DeleteController(context.Background(), controllerName); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err})
	}

	c.String(http.StatusNoContent, "")
}

func (router *ControllersRouteHandler) Update(c *gin.Context) {
	controllerName := c.Param("controller_name")
	body := struct {
		Name      string `json:"name" binding:"required"`
		Host      string `json:"host" binding:"required"`
		Readiness *bool  `json:"readiness" binding:"required"`
	}{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if err := router.store.UpdateController(context.Background(), controllerName, types.Controller{
		Host:      body.Host,
		Name:      body.Name,
		Readiness: *body.Readiness,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
}
