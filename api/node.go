package api

import (
	"net/http"

	"github.com/Dataman-Cloud/rolex/dockerclient"
	"github.com/gin-gonic/gin"
)

func GetNodes(c *gin.Context) {
	client, err := dockerclient.NewDockerGoClient("http://192.168.59.106:2376")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, err.Error())
		return
	}

	opts := dockerclient.NodeListOptions{}
	nodes, err := client.Nodelist(opts)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": nodes})
}
