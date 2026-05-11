package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		evt := log.Info()
		if len(c.Errors) > 0 {
			evt = log.Error().Strs("errors", c.Errors.Errors())
		}
		if rid, ok := c.Get(RequestIDKey); ok {
			evt = evt.Interface("request_id", rid)
		}
		evt.
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", raw).
			Int("status", c.Writer.Status()).
			Int("size", c.Writer.Size()).
			Str("ip", c.ClientIP()).
			Dur("latency", time.Since(start)).
			Msg("http request")
	}
}
