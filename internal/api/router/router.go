package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/jame-developer/aeontrac/internal/api/middleware"
	"github.com/jame-developer/aeontrac/internal/api/handlers"
)

func SetupRouter(logger *zap.Logger) *gin.Engine {
	r := gin.New()

	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.RecoveryMiddleware())

	r.GET("/report", handlers.ReportHandler)
	r.POST("/start", handlers.StartHandler)
	r.POST("/stop", handlers.StopHandler)

	return r
}