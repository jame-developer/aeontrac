package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jame-developer/aeontrac/internal/service"
)

func StatusHandler(c *gin.Context) {
	isTracking := service.IsTracking()
	c.JSON(http.StatusOK, gin.H{"is_tracking": isTracking})
}
