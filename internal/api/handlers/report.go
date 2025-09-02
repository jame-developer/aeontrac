package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jame-developer/aeontrac/configuration"
	"github.com/jame-developer/aeontrac/pkg/models"
	"go.uber.org/zap"

	"github.com/jame-developer/aeontrac/pkg/reporting"
)

func AttachReportingHandlers(grp *gin.RouterGroup, logger *zap.Logger, config *configuration.Config, data *models.AeonVault, folder string) {
	grp.GET("/report", reportHandler(config, data))
	grp.GET("/report/today", todayReportHandler(config, data))
}
func reportHandler(config *configuration.Config, data *models.AeonVault) func(c *gin.Context) {
	return func(c *gin.Context) {
		fromStr := c.DefaultQuery("from", time.Now().Format(time.DateOnly))
		toStr := c.DefaultQuery("to", time.Now().Format(time.DateOnly))

		from, err := time.Parse(time.DateOnly, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from date format"})
			return
		}

		to, err := time.Parse(time.DateOnly, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to date format"})
			return
		}

		report := reporting.GetReport(from, to, config.WorkingHours, data)

		c.JSON(http.StatusOK, report)
	}
}
func todayReportHandler(config *configuration.Config, data *models.AeonVault) func(c *gin.Context) {
	return func(c *gin.Context) {
		todayReport := reporting.GetTodayReport(config.WorkingHours, data)
		c.JSON(http.StatusOK, todayReport)
	}
}
