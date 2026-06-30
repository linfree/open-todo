package router

import (
	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/server/internal/database"
	"github.com/linfree/open-todo/server/internal/handler"
	"github.com/linfree/open-todo/server/internal/middleware"
)

func New(db *database.DB, jwtSecret string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	authH := &handler.AuthHandler{DB: db, JWTSecret: jwtSecret}
	syncH := &handler.SyncHandler{DB: db}

	r.GET("/api/v1/health", handler.Health)

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
	}

	sync := r.Group("/api/v1/sync")
	sync.Use(middleware.AuthRequired(jwtSecret))
	{
		sync.POST("/push", syncH.Push)
		sync.POST("/pull", syncH.Pull)
	}

	return r
}
