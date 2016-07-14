package api

import (
	"net/http"

	"github.com/Dataman-Cloud/rolex/util"

	goclient "github.com/fsouza/go-dockerclient"
	"github.com/gin-gonic/gin"
)

func (api *Api) InspectVolume(ctx *gin.Context) {
	volume, err := api.GetDockerClient().InspectVolume(ctx.Param("node_id"), ctx.Param("name"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": util.ENGINE_OPERATION_ERROR, "data": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": util.OPERATION_SUCCESS, "data": volume})
}

func (api *Api) ListVolume(ctx *gin.Context) {
	var opts goclient.ListVolumesOptions

	/*if err := ctx.BindJSON(&opts); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": util.PARAMETER_ERROR, "data": err.Error()})
		return
	}*/

	volumes, err := api.GetDockerClient().ListVolumes(ctx.Param("node_id"), opts)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": util.ENGINE_OPERATION_ERROR, "data": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": util.OPERATION_SUCCESS, "data": volumes})
}

func (api *Api) CreateVolume(ctx *gin.Context) {
	var opts goclient.CreateVolumeOptions

	if err := ctx.BindJSON(&opts); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": util.PARAMETER_ERROR, "data": err.Error()})
		return
	}

	volume, err := api.GetDockerClient().CreateVolume(ctx.Param("node_id"), opts)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": util.ENGINE_OPERATION_ERROR, "data": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"code": util.OPERATION_SUCCESS, "data": volume})
}

func (api *Api) RemoveVolume(ctx *gin.Context) {
	if err := api.GetDockerClient().RemoveVolume(ctx.Param("node_id"), ctx.Param("name")); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": util.ENGINE_OPERATION_ERROR, "data": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": util.OPERATION_SUCCESS, "data": "remove success"})
}
