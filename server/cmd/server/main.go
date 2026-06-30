package main

import (
	"log"

	"github.com/linfree/open-todo/server/internal/config"
	"github.com/linfree/open-todo/server/internal/database"
	"github.com/linfree/open-todo/server/internal/router"
)

func main() {
	cfg := config.Load()
	db, err := database.New(cfg.Driver, cfg.DSN)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	r := router.New(db, cfg.JWTSecret)
	log.Printf("open-todo server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
