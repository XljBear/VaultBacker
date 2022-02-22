package web

import (
	"VaultBacker/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func StartWebServer(config models.Config) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Vault server down...Sorry")
	})
	_ = r.Run(":" + strconv.Itoa(config.WebServer.Port))
}
