package handlers

import (
	"net/http"

	"aeontrac/internal/service"
	"aeontrac/pkg/models"
	"github.com/gin-gonic/gin"
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