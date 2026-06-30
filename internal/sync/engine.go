package sync

import (
	"log"
	"time"

	"github.com/linfree/open-todo/internal/store"
)

type Engine struct {
	store    *store.Store
	client   *Client
	lastSync time.Time
	stopCh   chan struct{}
}

func NewEngine(st *store.Store, serverURL, token string) *Engine {
	return &Engine{
		store:    st,
		client:   NewClient(serverURL, token),
		lastSync: time.Now().UTC(),
		stopCh:   make(chan struct{}),
	}
}

func (e *Engine) Start() {
	log.Println("[sync] engine started")

	go e.pushLoop()
	go e.pullLoop()
}

func (e *Engine) Stop() {
	close(e.stopCh)
}

func (e *Engine) pushLoop() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.doPush()
		}
	}
}

func (e *Engine) pullLoop() {
	// 启动时立即拉取
	e.doPull()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.doPull()
		}
	}
}

func (e *Engine) doPush() {
	changes, err := e.store.GetUnsyncedChanges()
	if err != nil || len(changes) == 0 {
		return
	}

	var records []SyncRecord
	var ids []int64
	for _, c := range changes {
		records = append(records, SyncRecord{
			TableName: c.TableName,
			RecordID:  c.RecordID,
			Action:    c.Action,
			Timestamp: c.Timestamp,
		})
		ids = append(ids, c.ID)
	}

	if err := e.client.Push(records); err != nil {
		log.Printf("[sync] push error: %v", err)
		return
	}

	e.store.MarkChangesSynced(ids)
	log.Printf("[sync] pushed %d changes", len(records))
}

func (e *Engine) doPull() {
	resp, err := e.client.Pull(e.lastSync.Format(time.RFC3339))
	if err != nil {
		log.Printf("[sync] pull error: %v", err)
		return
	}

	if len(resp.Changes) == 0 {
		return
	}

	for _, c := range resp.Changes {
		// TODO Phase 4: LWW merge into local DB
		log.Printf("[sync] pulled %s/%s", c.TableName, c.RecordID)
	}

	if t, err := time.Parse(time.RFC3339, resp.ServerTime); err == nil {
		e.lastSync = t
	}

	log.Printf("[sync] pulled %d changes", len(resp.Changes))
}
