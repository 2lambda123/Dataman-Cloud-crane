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
		v1.GET("/nodes", api.ListNodes)
		v1.GET("/nodes/:node_id", api.InspectNode)
		v1.GET("/nodes/:node_id/info", api.Info)

		// Going to delegate to /nodes/:id
		// v1.GET("/nodes/leader_manager", api.LeaderNode)

		v1.GET("/nodes/:node_id/containers", api.ListContainers)
		v1.GET("/nodes/:node_id/containers/:container_id", api.InspectContainer)
		v1.GET("/nodes/:node_id/containers/:container_id/diff", api.DiffContainer)
		v1.DELETE("/nodes/:node_id/containers/:container_id", api.DeleteContainer)
		v1.GET("/nodes/:node_id/containers/:container_id/logs", api.Logs)
		v1.GET("/nodes/:node_id/containers/:container_id/stats", api.Stats)
		//v1.GET("/logs/:node_id/:container_id", api.Logs)
		//v1.GET("/stats/:node_id/:container_id", api.Stats)

		v1.POST("/networks", api.CreateNetwork)
		v1.GET("/networks", api.ListNetworks)
		v1.DELETE("/networks/:network_id", api.RemoveNetwork)
		v1.GET("/networks/:network_id", api.InspectNetwork)
		v1.PATCH("/networks/:network_id", api.ConnectNetwork)

		v1.POST("/stacks", api.CreateStack)
		v1.GET("/stacks", api.ListStack)
		v1.GET("/stacks/:namespace", api.InspectStack)
		v1.DELETE("/stacks/:namespace", api.RemoveStack)
		v1.PATCH("/stacks/:namespace/services/:service_id", api.ScaleService)
		v1.GET("/stacks/:namespace/services", api.ListStackService)
		v1.GET("/stacks/:namespace/services/:service_id/tasks", api.ListTasks)
		v1.GET("/stacks/:namespace/services/:service_id/tasks/:task_id", api.InspectTask)

		v1.DELETE("/volumes/:node_id/:name", api.RemoveVolume)
		v1.GET("/volumes/:node_id/:name", api.InspectVolume)
		v1.GET("/volumes/:node_id", api.ListVolume)
		v1.POST("/volumes/:node_id", api.CreateVolume)

		v1.GET("/images/:node_id", api.ListImages)
		v1.GET("/images/:node_id/:image_id", api.InspectImage)
		v1.GET("/images/:node_id/:image_id/history", api.ImageHistory)

	}

	misc := router.Group("/misc/v1")
	{
		misc.GET("/help", api.Help(router))
		misc.GET("/config", api.RolexConfig)
		misc.GET("/health", api.HealthCheck)
	}

	return router
}
