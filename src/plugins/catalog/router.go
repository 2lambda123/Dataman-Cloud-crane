package catalog

import (
	"github.com/gin-gonic/gin"
)

func (catalogApi *CatalogApi) RegisterApiForCatalog(router *gin.Engine, middlewares ...gin.HandlerFunc) {

	catalogV1 := router.Group("/catalog/v1", middlewares...)
	{
		catalogV1.GET("/catalogs", catalogApi.ListCatalog)
		catalogV1.GET("/catalogs/:name", catalogApi.GetCatalog)
		catalogV1.Static("/icons", catalogApi.Config.CatalogPath)
	}
}
