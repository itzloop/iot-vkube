package utils

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{
			"method":      r.Method,
			"remote_addr": r.RemoteAddr,
			"uri":         r.RequestURI,
		}).Info("new request")

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func GetEntryFromContext(ctx context.Context) *logrus.Entry {
	v := ctx.Value("entry")
	entry, ok := v.(*logrus.Entry)
	if !ok {
		return logrus.NewEntry(logrus.StandardLogger())
	}

	return entry
}

func ContextWithSpot(ctx context.Context, spot string) context.Context {
	entry := GetEntryFromContext(ctx)
	return ContextWithEntry(ctx, entry.WithField("spot", spot))
}

func ContextWithEntry(ctx context.Context, entry *logrus.Entry) context.Context {
	if entry == nil {
		entry = logrus.NewEntry(logrus.StandardLogger())
	}

	return context.WithValue(ctx, "entry", entry)
}

func WaitWithThreeDots(msg string, delay time.Duration) {
	start := time.Now()
	ticker := time.NewTicker(time.Millisecond * 250)
	defer ticker.Stop()

	dots := 0
	maxDots := 3
	for time.Since(start) <= delay {
		select {
		case <-ticker.C:
			dots = (dots + 1) % (maxDots + 1)
			fmt.Printf("%s\r", strings.Repeat(" ", maxDots+len(msg)))
			fmt.Printf("%s%s\r", msg, strings.Repeat(".", dots))
		}
	}
}
