package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/benauro/kube-cdn/cdn/handler"
	"github.com/benauro/kube-cdn/cdn/logger"
	"github.com/benauro/kube-cdn/cdn/middleware"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithFormatter(logger.Format))
	r.Use(middleware.RequestLoggerMiddleware)

	auth := r.Group("/", gin.BasicAuth(gin.Accounts{"root": "123"}))
	{
		auth.GET("/")
	}

	r.GET("/cdn/media/:mediaID", handler.GetMedia)

	log.Printf("Start serving at: %v", ":8080")
	r.Run(":8080")
}

func GetContentType(ext string) string {
	switch ext {
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".pdf":
		return "application/pdf"
	case ".ps":
		return "application/postscript"
	default:
		return ""
	}
}
