package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/linfree/open-todo/internal/config"
	"github.com/linfree/open-todo/internal/server"
	"github.com/linfree/open-todo/internal/store"
	"github.com/linfree/open-todo/internal/sync"
	"github.com/linfree/open-todo/internal/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	st, err := store.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	if err := st.InitDefaults(); err != nil {
		log.Printf("init defaults: %v", err)
	}

	// 启动 SyncEngine (如果配置了服务端)
	if cfg.ServerURL != "" && cfg.AuthToken != "" {
		engine := sync.NewEngine(st, cfg.ServerURL, cfg.AuthToken)
		engine.Start()
		defer engine.Stop()
	}

	router := server.New(st, cfg)
	registerStaticRoutes(router)

	addr := fmt.Sprintf(":%d", cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}

	log.Printf("open-todo server starting on http://localhost%s", addr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("shutting down...")
		listener.Close()
	}()

	go func() {
		if err := router.RunListener(listener); err != nil {
			log.Printf("server: %v", err)
		}
	}()

	uiApp := ui.New(cfg.Port)
	uiApp.Run(func() {
		log.Println("UI ready")
	})

	select {}
}
