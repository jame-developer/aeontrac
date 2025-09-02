package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jame-developer/aeontrac/configuration"
	"github.com/jame-developer/aeontrac/internal/service"
	"github.com/jame-developer/aeontrac/pkg/models"
	"github.com/jame-developer/aeontrac/pkg/repositories"
	"github.com/jame-developer/aeontrac/pkg/tracking"
	"go.uber.org/zap"
)

func AttachTrackingHandlers(grp *gin.RouterGroup, logger *zap.Logger, config *configuration.Config, data *models.AeonVault, folder string) {
	grp.POST("/tracking/start", startHandler(logger, data, folder))
	grp.POST("/tracking/stop", stopHandler(logger, config, data, folder))
	grp.GET("/tracking/status", statusHandler())
}

func statusHandler() func(c *gin.Context) {
	return func(c *gin.Context) {
		isTracking := service.IsTracking()
		c.JSON(http.StatusOK, gin.H{"is_tracking": isTracking})
	}
}

func startHandler(logger *zap.Logger, data *models.AeonVault, dataFolder string) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req StartTrackingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Invalid request body", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := tracking.StartTracking(nil, req.Comment, data); err != nil {
			logger.Error("Error starting tracking", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error starting tracking"})
			return
		}
		if err := repositories.SaveAeonVault(dataFolder, *data); err != nil {
			logger.Error("Failed to save app", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save app"})
			return
		}

		c.String(http.StatusOK, "Time tracking started successfully.")
	}
}
func stopHandler(logger *zap.Logger, config *configuration.Config, data *models.AeonVault, dataFolder string) func(c *gin.Context) {
	return func(c *gin.Context) {
		if err := tracking.StopTracking(nil, config.WorkingHours, data); err != nil {
			logger.Error("Error starting tracking", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error starting tracking"})
			return
		}
		if err := repositories.SaveAeonVault(dataFolder, *data); err != nil {
			logger.Error("Failed to save app", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save app"})
			return
		}

		c.String(http.StatusOK, "Time tracking stopped successfully.")
	}
}
