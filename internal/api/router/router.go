package router

import (
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/jame-developer/aeontrac/configuration"
	"github.com/jame-developer/aeontrac/pkg/models"
	"go.uber.org/zap"

	"github.com/jame-developer/aeontrac/internal/api/handlers"
	"github.com/jame-developer/aeontrac/internal/api/middleware"
)

func SetupRouter(logger *zap.Logger, config *configuration.Config, data *models.AeonVault, dataFolder string, webTemplateFolder string) *gin.Engine {
	r := gin.New()

	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.RecoveryMiddleware())

	apiGrp := r.Group("/api")
	handlers.AttachReportingHandlers(apiGrp, logger, config, data, dataFolder)
	handlers.AttachTrackingHandlers(apiGrp, logger, config, data, dataFolder)

	r.POST("/worktime", handlers.AddWorkTimeHandler)

	r.LoadHTMLGlob(path.Join(webTemplateFolder, "web/templates/*"))
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.StaticFS("/doc", http.Dir("web/static"))

	return r
}
