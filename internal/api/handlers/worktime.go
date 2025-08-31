package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jame-developer/aeontrac/internal/service"
	"github.com/jame-developer/aeontrac/pkg/models"
)

// AddWorkTimeHandler handles the addition of a new work time entry.
func AddWorkTimeHandler(c *gin.Context) {
	var req models.WorkTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	aeonUnit, err := service.AddWorkTimeEntry(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, aeonUnit)
}
