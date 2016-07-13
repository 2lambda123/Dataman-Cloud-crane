package api

import (
	"time"

	"github.com/Dataman-Cloud/rolex/api/middlewares"
	"github.com/Dataman-Cloud/rolex/plugins/registry"
	"github.com/Dataman-Cloud/rolex/util/log"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

func (api *Api) ApiRouter() *gin.Engine {
	router := gin.New()

	router.Use(log.Ginrus(logrus.StandardLogger(), time.RFC3339, true), gin.Recovery())
	router.Use(middlewares.OptionHandler())

	router.GET("/", func(c *gin.Context) {
		c.String(200, "pass")
	})

	if api.Config.FeatureEnabled("registry") {
		r := &registry.Registry{Config: api.Config}
		r.RegisterApiForRegistry(router)
	}

	v1 := router.Group("/api/v1", middlewares.Authorization)
	{
		v1.GET("/health", api.HealthCheck)
		v1.GET("/nodes", api.ListNodes)
		v1.GET("/nodes/:id", api.InspectNode)

		v1.GET("/containers", api.ListContainers)
		v1.GET("/containers/:id", api.InspectContainer)

		v1.POST("/networks", api.CreateNetwork)
		v1.GET("/networks", api.ListNetworks)
		v1.DELETE("/networks/:id", api.RemoveNetwork)
		v1.GET("/networks/:id", api.InspectNetwork)
		v1.PATCH("/networks/:id", api.ConnectNetwork)

		v1.POST("/stacks", api.CreateStack)
		v1.GET("/stacks", api.ListStack)
		v1.GET("/stacks/:namespace", api.InspectStack)
		v1.PATCH("/stacks/:namespace/services/:serviceID", api.ScaleService)
		v1.GET("/stacks/:namespace/services/:serviceI", api.ListStackService)
	}

	return router
}
