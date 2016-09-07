package catalog

import (
	//"io/ioutil"

	"github.com/Dataman-Cloud/go-component/utils/db"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

func NewCatalog(catalogPath string) *CatalogApi {
	return &CatalogApi{
		CatalogPath: catalogPath,
		DbClient:    db.DB(),
	}
}

func (catalogApi *CatalogApi) MigriateTable() {
	catalogApi.DbClient.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&Catalog{})
}

func (catalogApi *CatalogApi) RegisterApiForCatalog(router *gin.Engine, middlewares ...gin.HandlerFunc) {

	catalogV1 := router.Group("/catalog/v1", middlewares...)
	{
		catalogV1.GET("/catalogs", catalogApi.ListCatalog)
		catalogV1.POST("/catalogs", catalogApi.CreateCatalog)
		catalogV1.GET("/catalogs/:catalog_id", catalogApi.GetCatalog)
		catalogV1.DELETE("/catalogs/:catalog_id", catalogApi.DeleteCatalog)
		catalogV1.PATCH("/catalogs/:catalog_id", catalogApi.UpdateCatalog)
	}

	catalogApi.LoadCatalog()
}

func (catalogApi *CatalogApi) LoadCatalog() {
	catalogs, err := AllCatalogFromPath(catalogApi.CatalogPath)
	if err != nil {
		log.Errorf("load catalogs error: %v", err)
		return
	}

	for _, catalog := range catalogs {
		catalog.UserId = CATALOG_SYSTEM_DEFAULT
		catalog.Type = CATALOG_SYSTEM_DEFAULT
		catalogApi.Save(catalog)
	}
}
