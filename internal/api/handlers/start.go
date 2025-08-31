package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/jame-developer/aeontrac/internal/api/middleware"
	"github.com/jame-developer/aeontrac/internal/appcore"
	"github.com/jame-developer/aeontrac/pkg/commands"
)

func StartHandler(c *gin.Context) {
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

	var req TimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	config, data, dataFolder, err := appcore.LoadApp()
	if err != nil {
		logger.Error("Failed to load app", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load app"})
		return
	}

	args := []string{}
	if req.Time != nil && *req.Time != "" {
		args = append(args, *req.Time)
	}

	commands.StartCommand(args, data)
	if err := appcore.SaveApp(config, data, dataFolder); err != nil {
		logger.Error("Failed to save app", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save app"})
		return
	}

	c.String(http.StatusOK, "Time tracking started successfully.")
}
