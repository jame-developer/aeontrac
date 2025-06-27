package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LoggerMiddleware returns a Gin middleware that logs requests using zap.Logger
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Generate unique request ID
		reqID := uuid.New().String()

		// Add logger and request ID to context
		c.Set(LoggerKey, logger)
		c.Set(RequestIDKey, reqID)

		// Process request
		c.Next()

		// After request
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		logger.Info("request",
			zap.String("request_id", reqID),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
		)
	}
}