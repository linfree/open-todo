package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/server/internal/database"
	"github.com/linfree/open-todo/server/internal/handler"
	"github.com/linfree/open-todo/server/internal/middleware"
)

func New(db database.Database, jwtSecret string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	authH := &handler.AuthHandler{DB: db, JWTSecret: jwtSecret}
	syncH := &handler.SyncHandler{DB: db}
	catH := &handler.CategoryHandler{DB: db}
	tagH := &handler.TagHandler{DB: db}

	r.GET("/api/v1/health", handler.Health)

	rateLimiter := middleware.NewRateLimiter(20, time.Minute)
	auth := r.Group("/api/v1/auth")
	auth.Use(rateLimiter.Middleware())
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
	}

	apiGroup := r.Group("/api/v1")
	apiGroup.Use(middleware.AuthRequired(jwtSecret))

	categories := apiGroup.Group("/categories")
	{
		categories.GET("", catH.GetCategories)
		categories.POST("", catH.SaveCategory)
		categories.DELETE("/:id", catH.DeleteCategory)
	}

	tags := apiGroup.Group("/tags")
	{
		tags.GET("", tagH.GetTags)
		tags.POST("", tagH.SaveTag)
		tags.DELETE("/:id", tagH.DeleteTag)
	}

	sync := r.Group("/api/v1/sync")
	sync.Use(middleware.AuthRequired(jwtSecret))
	{
		sync.POST("/push", syncH.Push)
		sync.POST("/pull", syncH.Pull)
	}

	return r
}
