// Package middleware provides Gin middleware for the API.
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RequestIDHeader = "X-Request-ID"
	requestIDKey    = "request_id"
)


func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" || len(id) > 128 {
			id = newID()
		}
		c.Set(requestIDKey, id)
		c.Header(RequestIDHeader, id)
		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	return c.GetString(requestIDKey)
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("%s %s %d %dB %s id=%s",
			c.Request.Method, c.Request.URL.RequestURI(), c.Writer.Status(), c.Writer.Size(),
			time.Since(start).Round(time.Microsecond), GetRequestID(c))
	}
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		log.Printf("panic: %v id=%s", err, GetRequestID(c))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	})
}
