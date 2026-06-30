package app

import (
	"log"
	"time"

	"github.com/linfree/open-todo/internal/store"
)

func StartReminderChecker(st *store.Store, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			// TODO: Phase 4 完整实现提醒逻辑
			log.Println("[reminder] check running...")
			_ = st
		}
	}()
}
