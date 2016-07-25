package api

import (
	"time"

	"github.com/Dataman-Cloud/rolex/api/middlewares"
	"github.com/Dataman-Cloud/rolex/plugins/auth"
	"github.com/Dataman-Cloud/rolex/plugins/auth/authenticators"
	chains "github.com/Dataman-Cloud/rolex/plugins/auth/middlewares"
	"github.com/Dataman-Cloud/rolex/plugins/auth/token_store"
	"github.com/Dataman-Cloud/rolex/plugins/registry"
	"github.com/Dataman-Cloud/rolex/util/log"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

func (api *Api) ApiRouter() *gin.Engine {
	router := gin.New()
	Authorization := middlewares.Authorization
	AuthorizeServiceAccess := middlewares.AuthorizeServiceAccess

	router.Use(log.Ginrus(logrus.StandardLogger(), time.RFC3339, true), gin.Recovery())
	router.Use(middlewares.OptionHandler())
	router.Use(middlewares.RolexApiContext())

	router.GET("/", func(c *gin.Context) {
		c.String(200, "pass")
	})

	if api.Config.FeatureEnabled("registry") {
		r := &registry.Registry{Config: api.Config}
		r.RegisterApiForRegistry(router)
	}

	if api.Config.FeatureEnabled("account") {
		a := &auth.AccountApi{Config: api.Config, RolexDockerClient: api.Client}
		if api.Config.AccountTokenStore == "default" {
			a.TokenStore = token_store.NewDefaultStore()
		}

		if api.Config.AccountAuthenticator == "default" {
			a.Authenticator = authenticators.NewDefaultAuthenticator()
		} else if api.Config.AccountAuthenticator == "db" {
			a.Authenticator = authenticators.NewDBAuthenticator()
		}

		Authorization = chains.Authorization(a)
		AuthorizeServiceAccess = chains.AuthorizeServiceAccess(a)

		// account mode, Authorization enabled
		authorizeMiddlewares := make(map[string](func(permissionRequired auth.Permission) gin.HandlerFunc), 0)
		authorizeMiddlewares["AuthorizeServiceAccess"] = AuthorizeServiceAccess
		a.RegisterApiForAccount(router, authorizeMiddlewares, chains.Authorization(a), middlewares.ListIntercept())
	}

	v1 := router.Group("/api/v1", Authorization, middlewares.ListIntercept())
	{
		v1.GET("/nodes", api.ListNodes)
		v1.GET("/nodes/:node_id", api.InspectNode)
		v1.GET("/nodes/:node_id/info", api.Info)
		// Going to delegate to /nodes/:id
		// v1.GET("/nodes/leader_manager", api.LeaderNode)

		// Containers
		v1.GET("/nodes/:node_id/containers/:container_id/terminal", api.ConnectContainer)
		v1.GET("/nodes/:node_id/containers", api.ListContainers)
		v1.GET("/nodes/:node_id/containers/:container_id", api.InspectContainer)
		v1.GET("/nodes/:node_id/containers/:container_id/diff", api.DiffContainer)
		v1.DELETE("/nodes/:node_id/containers/:container_id", api.DeleteContainer)
		v1.GET("/nodes/:node_id/containers/:container_id/logs", api.LogsContainer)
		v1.GET("/nodes/:node_id/containers/:container_id/stats", api.StatsContainer)

		// Images
		v1.GET("/nodes/:node_id/images", api.ListImages)
		v1.GET("/nodes/:node_id/images/:image_id", api.InspectImage)
		v1.GET("/nodes/:node_id/images/:image_id/history", api.ImageHistory)
		v1.DELETE("/nodes/:node_id/images/:image_id", api.RemoveImage)

		// Volumes
		v1.GET("/nodes/:node_id/volumes", api.ListVolume)
		v1.GET("/nodes/:node_id/volumes/:volume_id", api.InspectVolume)
		v1.POST("/nodes/:node_id/volumes", api.CreateVolume)
		v1.DELETE("/nodes/:node_id/volumes/:volume_id", api.RemoveVolume)
		v1.PATCH("/nodes/:node_id", api.UpdateNode)

		// Networks
		v1.POST("/nodes/:node_id/networks", api.CreateNodeNetwork)
		v1.GET("/nodes/:node_id/networks", api.ListNodeNetworks)
		v1.GET("/nodes/:node_id/networks/:network_id", api.InspectNodeNetwork)
		v1.PATCH("/nodes/:node_id/networks/:network_id", api.ConnectNodeNetwork)

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
		v1.GET("/stacks/:namespace/services/:service_id", AuthorizeServiceAccess(auth.PermReadOnly), api.InspectService)
		v1.GET("/stacks/:namespace/services", AuthorizeServiceAccess(auth.PermReadOnly), api.ListStackService)
		v1.GET("/stacks/:namespace/services/:service_id/logs", api.LogsService)
		v1.GET("/stacks/:namespace/services/:service_id/stats", api.StatsService)
		v1.GET("/stacks/:namespace/services/:service_id/tasks", api.ListTasks)
		v1.GET("/stacks/:namespace/services/:service_id/tasks/:task_id", api.InspectTask)
	}

	misc := router.Group("/misc/v1")
	{
		misc.GET("/help", api.Help(router))
		misc.GET("/config", api.RolexConfig)
		misc.GET("/health", api.HealthCheck)
	}

	return router
}
