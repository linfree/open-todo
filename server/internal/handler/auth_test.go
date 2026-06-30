package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/linfree/open-todo/server/internal/database"
)

func setupAuthHandler(t *testing.T) (*AuthHandler, database.Database) {
	t.Helper()
	db, err := database.New("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &AuthHandler{DB: db, JWTSecret: "test-secret"}, db
}

type authResponse struct {
	User  *database.User `json:"user"`
	Token string         `json:"token"`
	Error string         `json:"error"`
}

func TestRegisterSuccess(t *testing.T) {
	h, _ := setupAuthHandler(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodPost, "/api/v1/auth/register",
		strings.NewReader(`{"email":"test@example.com","name":"Test User","password":"password123"}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp authResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if resp.User == nil {
		t.Fatal("expected non-nil user")
	}
	if resp.User.Email != "test@example.com" {
		t.Fatalf("expected email 'test@example.com', got '%s'", resp.User.Email)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	h, _ := setupAuthHandler(t)

	body := `{"email":"dup@example.com","name":"Dup User","password":"password123"}`

	// First registration - should succeed
	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	c1.Request.Header.Set("Content-Type", "application/json")
	h.Register(c1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("first registration: expected 201, got %d: %s", w1.Code, w1.Body.String())
	}

	// Second registration with same email - should fail
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	c2.Request.Header.Set("Content-Type", "application/json")
	h.Register(c2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d: %s", w2.Code, w2.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(w2.Body.Bytes(), &resp)
	if !strings.Contains(resp["error"], "already exists") {
		t.Fatalf("expected 'already exists' error, got '%s'", resp["error"])
	}
}

func TestLoginSuccess(t *testing.T) {
	h, _ := setupAuthHandler(t)

	// Register first
	regBody := `{"email":"login@example.com","name":"Login User","password":"password123"}`
	wReg := httptest.NewRecorder()
	cReg, _ := gin.CreateTestContext(wReg)
	cReg.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(regBody))
	cReg.Request.Header.Set("Content-Type", "application/json")
	h.Register(cReg)

	if wReg.Code != http.StatusCreated {
		t.Fatalf("registration failed: expected 201, got %d", wReg.Code)
	}

	// Now login
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"email":"login@example.com","password":"password123"}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp authResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if resp.User == nil {
		t.Fatal("expected non-nil user")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	h, _ := setupAuthHandler(t)

	// Register first
	regBody := `{"email":"wrong@example.com","name":"Wrong User","password":"password123"}`
	wReg := httptest.NewRecorder()
	cReg, _ := gin.CreateTestContext(wReg)
	cReg.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(regBody))
	cReg.Request.Header.Set("Content-Type", "application/json")
	h.Register(cReg)

	if wReg.Code != http.StatusCreated {
		t.Fatalf("registration failed: expected 201, got %d", wReg.Code)
	}

	// Login with wrong password
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodPost, "/api/v1/auth/login",
		strings.NewReader(`{"email":"wrong@example.com","password":"wrongpassword"}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}
