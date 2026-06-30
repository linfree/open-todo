package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/linfree/open-todo/server/internal/database"
	"github.com/linfree/open-todo/server/internal/router"
)

const testJWTSecret = "test-sync-secret"

func setupSyncUser(t *testing.T) (database.Database, string) {
	t.Helper()
	db, err := database.New("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	user := &database.User{
		ID:        "sync-user-1",
		Email:     "sync@example.com",
		Name:      "Sync User",
		Password:  "$2a$10$dummyhashedpasswordhere123456789012345678901234567890", // bcrypt hash placeholder
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if _, err := db.CreateUser(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return db, generateTestToken(user.ID)
}

func generateTestToken(userID string) string {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testJWTSecret))
	return signed
}

func TestPushReturns200WithServerTime(t *testing.T) {
	db, token := setupSyncUser(t)
	r := router.New(db, testJWTSecret)

	body := `{"changes":[]}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/push", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if _, ok := resp["server_time"]; !ok {
		t.Fatal("expected server_time in response")
	}
}

func TestPushWithoutTokenReturns401(t *testing.T) {
	db, _ := setupSyncUser(t)
	r := router.New(db, testJWTSecret)

	body := `{"changes":[]}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/push", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestPullReturns200WithChangesArray(t *testing.T) {
	db, token := setupSyncUser(t)
	r := router.New(db, testJWTSecret)

	// First push a change so pull has something to return
	pushBody := `{
		"changes": [
			{
				"table_name": "tasks",
				"record_id": "task-1",
				"action": "insert",
				"timestamp": "2024-01-01T00:00:00Z",
				"data": {
					"id": "task-1",
					"title": "Test Task",
					"completed": false,
					"priority": "none",
					"status": "todo",
					"list_id": "list-1",
					"tags": "[]",
					"sub_tasks": "[]",
					"reminders": "[]",
					"deleted": false,
					"order_num": 0
				}
			}
		]
	}`

	wPush := httptest.NewRecorder()
	reqPush := httptest.NewRequest(http.MethodPost, "/api/v1/sync/push", strings.NewReader(pushBody))
	reqPush.Header.Set("Content-Type", "application/json")
	reqPush.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(wPush, reqPush)

	if wPush.Code != http.StatusOK {
		t.Fatalf("push failed: expected 200, got %d: %s", wPush.Code, wPush.Body.String())
	}

	// Now pull with since=epoch
	pullBody := `{"since":"1970-01-01T00:00:00Z"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/pull", strings.NewReader(pullBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	changes, ok := resp["changes"]
	if !ok {
		t.Fatal("expected changes in response")
	}

	changesArr, ok := changes.([]interface{})
	if !ok {
		t.Fatalf("expected changes to be an array, got %T", changes)
	}
	if len(changesArr) == 0 {
		t.Fatal("expected at least one change in response")
	}

	if _, ok := resp["server_time"]; !ok {
		t.Fatal("expected server_time in response")
	}
}

func TestPullWithoutTokenReturns401(t *testing.T) {
	db, _ := setupSyncUser(t)
	r := router.New(db, testJWTSecret)

	body := `{"since":"2024-01-01T00:00:00Z"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/pull", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}
