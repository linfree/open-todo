package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/linfree/open-todo/internal/config"
	"github.com/linfree/open-todo/internal/store"
)

func TestGetTasks(t *testing.T) {
	st, _ := store.Open(":memory:")
	defer st.Close()
	cfg := config.Default()
	router := New(st, cfg)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tasks", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var tasks []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&tasks)
}

func TestCreateTask(t *testing.T) {
	st, _ := store.Open(":memory:")
	defer st.Close()
	cfg := config.Default()
	router := New(st, cfg)

	body := strings.NewReader(`{"title":"API测试","list_id":"default"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
