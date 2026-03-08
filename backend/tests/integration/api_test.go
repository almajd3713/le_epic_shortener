//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"shortener.reeler.com/backend/internal/cache"
	"shortener.reeler.com/backend/internal/config"
	"shortener.reeler.com/backend/internal/db"
	"shortener.reeler.com/backend/internal/handlers"
	"shortener.reeler.com/backend/internal/middleware"
	"shortener.reeler.com/backend/internal/models"
	"shortener.reeler.com/backend/internal/repository"
	"shortener.reeler.com/backend/internal/server"
	"shortener.reeler.com/backend/internal/services"
)

const testBaseURL = "http://localhost:8080"

// suite holds shared infrastructure for all integration tests.
type suite struct {
	router     *gin.Engine
	cache      cache.ICache
	pool       *pgxpool.Pool
	closeRedis func()
}

// ts is the package-level test suite, initialised once in TestMain.
var ts *suite

func TestMain(m *testing.M) {
	var err error
	ts, err = newSuite()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: integration test setup failed: %v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	ts.teardown()
	os.Exit(code)
}

// newSuite connects to Postgres and Redis, runs migrations, then wires up the
// full Gin router exactly as the production server does.
func newSuite() (*suite, error) {
	gin.SetMode(gin.TestMode)

	databaseURL := testEnv("TEST_DATABASE_URL", "postgresql://shortener_user:password@localhost:5433/shortener")
	redisURL := testEnv("TEST_REDIS_URL", "redis://localhost:6379")

	ctx := context.Background()

	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}

	if err := db.RunMigrations(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrations: %w", err)
	}

	redisClient, err := cache.NewRedisClient(ctx, config.CacheConfig{
		URL:             redisURL,
		MaxRetries:      3,
		MinRetryBackoff: 100 * time.Millisecond,
		MaxRetryBackoff: time.Second,
	})
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("redis connect: %w", err)
	}

	cacheInst := cache.NewCache(redisClient)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	urlRepo := repository.NewURLRepository(pool)
	cacheSvc := services.NewCacheService(cacheInst, logger)
	urlSvc := services.NewURLService(urlRepo, cacheSvc, logger)
	shortenerSvc := services.NewShortenerService(urlRepo, cacheSvc, logger)
	redirectSvc := services.NewRedirectorService(urlSvc, cacheSvc, logger)

	urlHandler := handlers.NewURLHandler(urlSvc)
	shortenerHandler := handlers.NewShortenerHandler(shortenerSvc)
	redirectHandler := handlers.NewRedirectHandler(redirectSvc)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.BaseURL(testBaseURL))
	server.SetupRoutes(r, *urlHandler, *shortenerHandler, *redirectHandler)

	return &suite{
		router:     r,
		cache:      cacheInst,
		pool:       pool,
		closeRedis: func() { redisClient.Close() },
	}, nil
}

func (s *suite) teardown() {
	s.closeRedis()
	s.pool.Close()
}

// reset wipes all URL rows and flushes the cache so each test starts clean.
func (s *suite) reset(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	if _, err := s.pool.Exec(ctx, "DELETE FROM urls"); err != nil {
		t.Fatalf("reset: DELETE FROM urls: %v", err)
	}
	if err := s.cache.Flush(ctx); err != nil {
		t.Fatalf("reset: cache flush: %v", err)
	}
}

// do fires an HTTP request against the test router and returns the recorder.
func (s *suite) do(method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

// shorten is a test helper: POST /api/shorten and return the decoded response.
func (s *suite) shorten(t *testing.T, longURL string, expiresAt *time.Time) models.URLResponse {
	t.Helper()
	body := map[string]any{"long_url": longURL}
	if expiresAt != nil {
		body["expires_at"] = expiresAt
	}
	rr := s.do(http.MethodPost, "/api/shorten", body)
	if rr.Code != http.StatusOK {
		t.Fatalf("shorten helper: expected 200, got %d; body: %s", rr.Code, rr.Body)
	}
	var resp models.URLResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("shorten helper: decode: %v", err)
	}
	return resp
}

func testEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestPing(t *testing.T) {
	ts.reset(t)

	rr := ts.do(http.MethodGet, "/ping", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["message"] != "pong" {
		t.Errorf("expected message=pong, got %q", body["message"])
	}
}

// ── POST /api/shorten ────────────────────────────────────────────────────────

func TestShorten_ValidURL(t *testing.T) {
	ts.reset(t)

	rr := ts.do(http.MethodPost, "/api/shorten", map[string]any{
		"long_url": "https://example.com/some/path?q=1",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rr.Code, rr.Body)
	}
	var resp models.URLResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.ShortCode) != 8 {
		t.Errorf("expected ShortCode length 8, got %d (%q)", len(resp.ShortCode), resp.ShortCode)
	}
	if want := testBaseURL + "/" + resp.ShortCode; resp.ShortURL != want {
		t.Errorf("expected ShortURL=%q, got %q", want, resp.ShortURL)
	}
	if resp.CreatedAt == "" {
		t.Error("expected non-empty CreatedAt")
	}
}

func TestShorten_InvalidJSON(t *testing.T) {
	ts.reset(t)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString("{bad json}"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestShorten_WithFutureExpiry(t *testing.T) {
	ts.reset(t)

	future := time.Now().Add(24 * time.Hour)
	resp := ts.shorten(t, "https://example.com/expiry", &future)

	// URL should be reachable before expiry.
	rr := ts.do(http.MethodGet, "/"+resp.ShortCode, nil)
	if rr.Code != http.StatusFound {
		t.Fatalf("expected 302 for non-expired URL, got %d", rr.Code)
	}

	// expires_at must appear in the listing.
	rr = ts.do(http.MethodGet, "/api/urls", nil)
	var items []models.URLListItem
	json.NewDecoder(rr.Body).Decode(&items)
	for _, item := range items {
		if item.ShortCode == resp.ShortCode {
			if item.ExpiresAt == nil {
				t.Error("expected non-nil ExpiresAt in list")
			}
			return
		}
	}
	t.Error("shortened URL not found in list")
}

// ── GET /:code (redirect) ────────────────────────────────────────────────────

func TestRedirect_Found(t *testing.T) {
	ts.reset(t)

	const longURL = "https://example.com/redirect-target"
	resp := ts.shorten(t, longURL, nil)

	rr := ts.do(http.MethodGet, "/"+resp.ShortCode, nil)
	if rr.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d; body: %s", rr.Code, rr.Body)
	}
	if loc := rr.Header().Get("Location"); loc != longURL {
		t.Errorf("expected Location=%q, got %q", longURL, loc)
	}
}

func TestRedirect_NotFound(t *testing.T) {
	ts.reset(t)

	rr := ts.do(http.MethodGet, "/nonexistentcode", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// Second redirect is served from the cache populated by the first.
func TestRedirect_ServedFromCache(t *testing.T) {
	ts.reset(t)

	const longURL = "https://example.com/cache-hit"
	resp := ts.shorten(t, longURL, nil)

	// First request (may go to DB or return from the cache seeded by shorten).
	if rr := ts.do(http.MethodGet, "/"+resp.ShortCode, nil); rr.Code != http.StatusFound {
		t.Fatalf("first redirect: expected 302, got %d", rr.Code)
	}

	// Second request must come from cache and return the same destination.
	rr := ts.do(http.MethodGet, "/"+resp.ShortCode, nil)
	if rr.Code != http.StatusFound {
		t.Fatalf("cached redirect: expected 302, got %d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != longURL {
		t.Errorf("cached redirect: expected Location=%q, got %q", longURL, loc)
	}
}

// ── GET /geturl/:code ────────────────────────────────────────────────────────

func TestGetURL_Found(t *testing.T) {
	ts.reset(t)

	const longURL = "https://example.com/geturl-test"
	resp := ts.shorten(t, longURL, nil)

	rr := ts.do(http.MethodGet, "/geturl/"+resp.ShortCode, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rr.Code, rr.Body)
	}
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["long_url"] != longURL {
		t.Errorf("expected long_url=%q, got %q", longURL, body["long_url"])
	}
}

func TestGetURL_NotFound(t *testing.T) {
	ts.reset(t)

	rr := ts.do(http.MethodGet, "/geturl/doesnotexist", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ── GET /api/urls ────────────────────────────────────────────────────────────

func TestListURLs(t *testing.T) {
	ts.reset(t)

	ts.shorten(t, "https://example.com/a", nil)
	ts.shorten(t, "https://example.com/b", nil)

	rr := ts.do(http.MethodGet, "/api/urls", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rr.Code, rr.Body)
	}
	var items []models.URLListItem
	if err := json.NewDecoder(rr.Body).Decode(&items); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
	for _, item := range items {
		if item.ShortCode == "" || item.LongURL == "" || item.ShortURL == "" {
			t.Errorf("item has empty required field: %+v", item)
		}
		if want := testBaseURL + "/" + item.ShortCode; item.ShortURL != want {
			t.Errorf("ShortURL: expected %q, got %q", want, item.ShortURL)
		}
		if !item.IsActive {
			t.Errorf("newly created URL should be active: %+v", item)
		}
	}
}

func TestListURLs_Empty(t *testing.T) {
	ts.reset(t)

	rr := ts.do(http.MethodGet, "/api/urls", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rr.Code, rr.Body)
	}
	// Decode as nullable array — both null and [] are valid empty responses.
	var items []models.URLListItem
	if err := json.NewDecoder(rr.Body).Decode(&items); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

// ── PATCH /api/toggle/:code ──────────────────────────────────────────────────

func TestToggle_Deactivate(t *testing.T) {
	ts.reset(t)

	resp := ts.shorten(t, "https://example.com/deactivate", nil)

	rr := ts.do(http.MethodPatch, "/api/toggle/"+resp.ShortCode, map[string]any{"action": "deactivate"})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 on deactivate, got %d; body: %s", rr.Code, rr.Body)
	}
	var body map[string]string
	json.NewDecoder(rr.Body).Decode(&body)
	if body["message"] != "URL deactivated successfully" {
		t.Errorf("unexpected message: %q", body["message"])
	}

	// Redirect must fail while the URL is inactive.
	rr = ts.do(http.MethodGet, "/"+resp.ShortCode, nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after deactivate, got %d", rr.Code)
	}
}

func TestToggle_Activate(t *testing.T) {
	ts.reset(t)

	resp := ts.shorten(t, "https://example.com/activate", nil)
	ts.do(http.MethodPatch, "/api/toggle/"+resp.ShortCode, map[string]any{"action": "deactivate"})

	rr := ts.do(http.MethodPatch, "/api/toggle/"+resp.ShortCode, map[string]any{"action": "activate"})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 on activate, got %d; body: %s", rr.Code, rr.Body)
	}
	var body map[string]string
	json.NewDecoder(rr.Body).Decode(&body)
	if body["message"] != "URL activated successfully" {
		t.Errorf("unexpected message: %q", body["message"])
	}

	// Redirect must work again after re-activation.
	rr = ts.do(http.MethodGet, "/"+resp.ShortCode, nil)
	if rr.Code != http.StatusFound {
		t.Fatalf("expected 302 after re-activate, got %d", rr.Code)
	}
}

func TestToggle_InvalidAction(t *testing.T) {
	ts.reset(t)

	resp := ts.shorten(t, "https://example.com", nil)

	rr := ts.do(http.MethodPatch, "/api/toggle/"+resp.ShortCode, map[string]any{"action": "invalid"})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d; body: %s", rr.Code, rr.Body)
	}
}

func TestToggle_MalformedBody(t *testing.T) {
	ts.reset(t)

	resp := ts.shorten(t, "https://example.com", nil)

	req := httptest.NewRequest(http.MethodPatch, "/api/toggle/"+resp.ShortCode, bytes.NewBufferString("{bad json}"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	ts.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// ── DELETE /api/delete/:code ─────────────────────────────────────────────────

func TestDelete(t *testing.T) {
	ts.reset(t)

	resp := ts.shorten(t, "https://example.com/delete-test", nil)

	rr := ts.do(http.MethodDelete, "/api/delete/"+resp.ShortCode, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete, got %d; body: %s", rr.Code, rr.Body)
	}
	var body map[string]string
	json.NewDecoder(rr.Body).Decode(&body)
	if body["message"] != "URL deleted successfully" {
		t.Errorf("unexpected message: %q", body["message"])
	}

	// Redirect must 404 after deletion.
	rr = ts.do(http.MethodGet, "/"+resp.ShortCode, nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rr.Code)
	}

	// Deleted URL must not appear in the listing.
	rr = ts.do(http.MethodGet, "/api/urls", nil)
	var items []models.URLListItem
	json.NewDecoder(rr.Body).Decode(&items)
	for _, item := range items {
		if item.ShortCode == resp.ShortCode {
			t.Error("deleted URL still present in list")
		}
	}
}

// ── Full lifecycle ───────────────────────────────────────────────────────────

// TestFullLifecycle exercises every endpoint in sequence to verify the system
// as a whole: shorten → redirect → lookup → list → deactivate → activate → delete.
func TestFullLifecycle(t *testing.T) {
	ts.reset(t)

	const longURL = "https://example.com/lifecycle"

	// 1. Shorten
	resp := ts.shorten(t, longURL, nil)

	// 2. Redirect
	rr := ts.do(http.MethodGet, "/"+resp.ShortCode, nil)
	if rr.Code != http.StatusFound {
		t.Fatalf("step 2 redirect: expected 302, got %d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != longURL {
		t.Errorf("step 2 redirect: expected Location=%q, got %q", longURL, loc)
	}

	// 3. Lookup via /geturl
	rr = ts.do(http.MethodGet, "/geturl/"+resp.ShortCode, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("step 3 geturl: expected 200, got %d", rr.Code)
	}

	// 4. Present in list and active
	rr = ts.do(http.MethodGet, "/api/urls", nil)
	var items []models.URLListItem
	json.NewDecoder(rr.Body).Decode(&items)
	found := false
	for _, item := range items {
		if item.ShortCode == resp.ShortCode {
			found = true
			if !item.IsActive {
				t.Error("step 4 list: URL should be active")
			}
		}
	}
	if !found {
		t.Fatal("step 4 list: URL not found in listing")
	}

	// 5. Deactivate
	rr = ts.do(http.MethodPatch, "/api/toggle/"+resp.ShortCode, map[string]any{"action": "deactivate"})
	if rr.Code != http.StatusOK {
		t.Fatalf("step 5 deactivate: expected 200, got %d", rr.Code)
	}

	// 6. Redirect fails after deactivation
	rr = ts.do(http.MethodGet, "/"+resp.ShortCode, nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("step 6 redirect after deactivate: expected 404, got %d", rr.Code)
	}

	// 7. Re-activate
	rr = ts.do(http.MethodPatch, "/api/toggle/"+resp.ShortCode, map[string]any{"action": "activate"})
	if rr.Code != http.StatusOK {
		t.Fatalf("step 7 activate: expected 200, got %d", rr.Code)
	}

	// 8. Redirect works again
	rr = ts.do(http.MethodGet, "/"+resp.ShortCode, nil)
	if rr.Code != http.StatusFound {
		t.Fatalf("step 8 redirect after activate: expected 302, got %d", rr.Code)
	}

	// 9. Delete
	rr = ts.do(http.MethodDelete, "/api/delete/"+resp.ShortCode, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("step 9 delete: expected 200, got %d", rr.Code)
	}

	// 10. Gone from list
	rr = ts.do(http.MethodGet, "/api/urls", nil)
	items = nil
	json.NewDecoder(rr.Body).Decode(&items)
	for _, item := range items {
		if item.ShortCode == resp.ShortCode {
			t.Error("step 10 list: deleted URL still present")
		}
	}
}
