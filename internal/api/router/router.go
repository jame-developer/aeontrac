package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/jame-developer/aeontrac/internal/api/handlers"
	"github.com/jame-developer/aeontrac/internal/api/middleware"
)

func SetupRouter(logger *zap.Logger) *gin.Engine {
	r := gin.New()

	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.RecoveryMiddleware())

	r.GET("/report", handlers.ReportHandler)
	r.POST("/start", handlers.StartHandler)
	r.POST("/stop", handlers.StopHandler)
	r.POST("/worktime", handlers.AddWorkTimeHandler)

	r.StaticFS("/doc", http.Dir("web/static"))

	return r
}
