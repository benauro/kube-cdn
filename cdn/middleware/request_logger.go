package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLoggerMiddleware(c *gin.Context) {
	// Log request details
	fmt.Printf("====================== REQUEST =======================\n")
	fmt.Printf("Time        : %s\n", time.Now().Format(time.RFC1123))
	fmt.Printf("Method      : %s\n", c.Request.Method)
	fmt.Printf("URL         : %s\n", c.Request.URL.String())
	fmt.Printf("Headers     :\n")
	for key, value := range c.Request.Header {
		fmt.Printf("  %s: %s\n", key, value)
	}
	fmt.Printf("Client IP   : %s\n", c.ClientIP())
	fmt.Printf("======================================================\n")

	// Process request
	c.Next()

	// Log response details
	fmt.Printf("====================== RESPONSE ======================\n")
	fmt.Printf("Status Code : %d\n", c.Writer.Status())
	fmt.Printf("Headers     :\n")
	for key, value := range c.Writer.Header() {
		fmt.Printf("  %s: %s\n", key, value)
	}
	fmt.Printf("======================================================\n")
}
