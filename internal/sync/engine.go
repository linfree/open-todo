package sync

import (
	"encoding/json"
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
		data, err := e.store.GetFullRecord(c.TableName, c.RecordID)
		if err != nil {
			log.Printf("[sync] skip %s/%s: %v", c.TableName, c.RecordID, err)
			continue
		}
		records = append(records, SyncRecord{
			TableName: c.TableName,
			RecordID:  c.RecordID,
			Action:    c.Action,
			Timestamp: c.Timestamp,
			Data:      data,
		})
		ids = append(ids, c.ID)
	}

	if len(records) == 0 {
		return
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

	merged := 0
	for _, c := range resp.Changes {
		localData, _ := e.store.GetFullRecord(c.TableName, c.RecordID)
		shouldWrite := len(localData) == 0

		if !shouldWrite && len(c.Data) > 0 {
			remoteTime, err1 := time.Parse(time.RFC3339, c.Timestamp)
			var localObj map[string]interface{}
			if err2 := json.Unmarshal(localData, &localObj); err2 == nil {
				if updatedAt, ok := localObj["updated_at"].(string); ok {
					localTime, err3 := time.Parse(time.RFC3339, updatedAt)
					if err1 == nil && err3 == nil && !remoteTime.Before(localTime) {
						shouldWrite = true
					}
				}
			}
		}

		if shouldWrite && len(c.Data) > 0 {
			if err := e.store.PutFullRecord(c.TableName, c.RecordID, c.Data); err != nil {
				log.Printf("[sync] merge error %s/%s: %v", c.TableName, c.RecordID, err)
				continue
			}
			merged++
		}
	}

	if t, err := time.Parse(time.RFC3339, resp.ServerTime); err == nil {
		e.lastSync = t
	}

	log.Printf("[sync] pulled %d changes, merged %d", len(resp.Changes), merged)
}
