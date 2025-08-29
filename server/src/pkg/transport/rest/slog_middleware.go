package rest

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// SlogMiddleware creates a middleware that logs HTTP requests using slog
func SlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		// Determine log level based on status code
		var level slog.Level
		switch {
		case statusCode >= 500:
			level = slog.LevelError
		case statusCode >= 400:
			level = slog.LevelWarn
		case statusCode >= 300:
			level = slog.LevelInfo
		default:
			level = slog.LevelInfo
		}

		logger.Log(c.Request.Context(), level, "HTTP Request",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", statusCode),
			slog.String("client_ip", clientIP),
			slog.Duration("latency", latency),
			slog.Int("body_size", bodySize),
			slog.String("user_agent", c.Request.UserAgent()),
		)

		// Log errors if any occurred
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("Request error",
					slog.String("method", method),
					slog.String("path", path),
					slog.String("error", err.Error()),
					slog.Int("type", int(err.Type)),
				)
			}
		}
	}
}

// SlogRecoveryMiddleware creates a recovery middleware that logs panics using slog
func SlogRecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.String("client_ip", c.ClientIP()),
					slog.Any("panic", err),
				)

				// Return 500 error
				c.AbortWithStatus(500)
			}
		}()

		c.Next()
	}
}
