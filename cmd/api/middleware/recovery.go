package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware returns a Gin middleware that recovers from panics and logs them
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Get logger from context
				loggerIface, exists := c.Get(LoggerKey)
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

				// Log panic and stack trace
				logger.Error("panic recovered",
					zap.Any("panic", r),
					zap.ByteString("stack", debug.Stack()),
					zap.String("request_id", c.GetString(RequestIDKey)),
				)

				// Respond with 500 and JSON error
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}