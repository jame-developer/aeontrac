package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/jame-developer/aeontrac/internal/appcore"
	"github.com/jame-developer/aeontrac/pkg/reporting"
	"github.com/jame-developer/aeontrac/cmd/api/middleware"
)

func ReportHandler(c *gin.Context) {
	loggerIface, exists := c.Get(middleware.LoggerKey)
	var logger *zap.Logger
	if exists {
		if l, ok := loggerIface.(*zap.Logger); ok {
			logger = l
		}
	}
	if logger == nil {
		// fallback logger if not found in context
		logger, _ = zap.NewProduction()
	}

	config, data, _, err := appcore.LoadApp()
	if err != nil {
		logger.Error("Failed to load app", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load app"})
		return
	}

	report := reporting.GetTodayReport(config.WorkingHours, data)

	c.JSON(http.StatusOK, report)
}