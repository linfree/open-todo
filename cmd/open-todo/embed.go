package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed web-dist/*
var webAssets embed.FS

func registerStaticRoutes(r *gin.Engine) {
	sub, err := fs.Sub(webAssets, "web-dist")
	if err != nil {
		return
	}

	fileServer := http.FileServer(http.FS(sub))

	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "public, max-age=31536000")
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Try serving static file first
		if strings.HasPrefix(path, "/") {
			path = path[1:]
		}
		if path != "" && strings.Contains(path, ".") {
			f, err := sub.Open(path)
			if err == nil {
				f.Close()
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		// SPA fallback
		data, err := webAssets.ReadFile("web-dist/index.html")
		if err != nil {
			c.String(http.StatusNotFound, "Frontend not built. Run: cd web && npm run build")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
}
