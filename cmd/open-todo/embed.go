package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed web-dist/*
var webAssets embed.FS

func registerStaticRoutes(r *gin.Engine) {
	sub, err := fs.Sub(webAssets, "web-dist")
	if err != nil {
		// 开发模式: 没有 web-dist 目录, 跳过
		return
	}

	fileServer := http.FileServer(http.FS(sub))

	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "public, max-age=31536000")
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	r.NoRoute(func(c *gin.Context) {
		data, err := webAssets.ReadFile("web-dist/index.html")
		if err != nil {
			c.String(http.StatusNotFound, "Frontend not built. Run: cd web && npm run build")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})
}
