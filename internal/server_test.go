package internal_test

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	// "net"
	// "net/url"
	// "net/http/httputil"
	// "context"

	"github.com/egor-lukin/ssl-proxy/internal"
	"github.com/egor-lukin/ssl-proxy/internal/servers"
	"github.com/egor-lukin/ssl-proxy/internal/settings"
)

// func TestProxyHandler_Found(t *testing.T) {
// 	// Start a test backend server
// 	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte("Hello from backend"))
// 	}))
// 	defer backend.Close()

// 	backendURL, _ := url.Parse(backend.URL)
// 	backendHost, backendPort, _ := net.SplitHostPort(backendURL.Host)

// 	db := &sql.DB{}
// 	serversList := []servers.Server{
// 		{Domain: "known.example.com", SSLKey: "key", SSLCert: "cert"},
// 	}
// 	settingsObj := &settings.Settings{
// 		DestinationDomain: "fake.example.com",
// 	}

// 	// Create handler
// 	handler := internal.ProxyHandler(db, serversList, settingsObj)

// 	// Patch the default transport for the reverse proxy to resolve fake.example.com to our backend
// 	originalTransport := http.DefaultTransport
// 	defer func() { http.DefaultTransport = originalTransport }()
// 	http.DefaultTransport = &http.Transport{
// 		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
// 			if addr == "fake.example.com:443" || addr == "fake.example.com:80" {
// 				addr = net.JoinHostPort(backendHost, backendPort)
// 			}
// 			d := net.Dialer{}
// 			return d.DialContext(ctx, network, addr)
// 		},
// 	}

// 	// Create request with known host
// 	req := httptest.NewRequest("GET", "https://known.example.com/", nil)
// 	req.Host = "known.example.com"
// 	w := httptest.NewRecorder()

// 	handler(w, req)

// 	t.Logf("HTTP status: %d", w.Code)

// 	if w.Body.String() != "Hello from backend" {
// 		t.Errorf("Expected backend response, got %q", w.Body.String())
// 	}
// }

func TestProxyHandler_NotFound(t *testing.T) {
	// Setup dummy data
	db := &sql.DB{} // In a real test, use a real or mock DB
	serversList := []servers.Server{
		{Domain: "known.example.com", SSLKey: "key", SSLCert: "cert"},
	}
	settingsObj := &settings.Settings{
		DestinationDomain: "destination.example.com",
	}

	// Create handler
	handler := internal.ProxyHandler(db, serversList, settingsObj)

	// Create request with unknown host
	req := httptest.NewRequest("GET", "https://unknown.example.com/", nil)
	req.Host = "unknown.example.com"
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found, got %d", w.Code)
	}
}

func TestInsertServer_Success(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if err := servers.CreateTable(db); err != nil {
		t.Fatalf("failed to create servers table: %v", err)
	}

	serversList := []servers.Server{}
	settingsObj := &settings.Settings{
		Token: "testtoken",
	}

	handler := internal.InsertServer(db, &serversList, settingsObj)

	serverJSON := `{"Domain":"new.example.com","SSLKey":"key","SSLCert":"cert"}`
	req := httptest.NewRequest("POST", "/__internal/servers",
		strings.NewReader(serverJSON))
	req.Header.Set("Authorization", "Bearer testtoken")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Server created") {
		t.Errorf("Expected response body to contain 'Server created', got %q", w.Body.String())
	}

	serversFromDB, err := servers.Select(db)
	if err != nil {
		t.Fatalf("Failed to select servers from db: %v", err)
	}
	found := false
	for _, s := range serversFromDB {
		if s.Domain == "new.example.com" && s.SSLKey == "key" && s.SSLCert == "cert" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find new server in database, but did not")
	}
}

func TestUpdateServer_Success(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if err := servers.CreateTable(db); err != nil {
		t.Fatalf("failed to create servers table: %v", err)
	}

	// Insert initial server
	origServer := servers.Server{
		Domain:  "update.example.com",
		SSLKey:  "oldkey",
		SSLCert: "oldcert",
	}
	if err := servers.Insert(db, &origServer); err != nil {
		t.Fatalf("failed to insert initial server: %v", err)
	}

	serversList, err := servers.Select(db)
	if err != nil {
		t.Fatalf("failed to select servers: %v", err)
	}
	settingsObj := &settings.Settings{
		Token: "testtoken",
	}

	handler := internal.UpdateServer(db, &serversList, settingsObj)

	// Prepare update JSON
	updateJSON := `{"Domain":"update.example.com","SSLKey":"newkey","SSLCert":"newcert"}`
	req := httptest.NewRequest("PUT", "/__internal/servers",
		strings.NewReader(updateJSON))
	req.Header.Set("Authorization", "Bearer testtoken")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Server stored") {
		t.Errorf("Expected response body to contain 'Server stored', got %q", w.Body.String())
	}

	// Check that the server was updated in the database
	serversFromDB, err := servers.Select(db)
	if err != nil {
		t.Fatalf("Failed to select servers from db: %v", err)
	}
	found := false
	for _, s := range serversFromDB {
		if s.Domain == "update.example.com" && s.SSLKey == "newkey" && s.SSLCert == "newcert" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find updated server in database, but did not")
	}
}

func TestHealthCheck_Success(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if err := servers.CreateTable(db); err != nil {
		t.Fatalf("failed to create servers table: %v", err)
	}

	serversList := []servers.Server{}
	settingsObj := &settings.Settings{
		Token: "testtoken",
	}

	handler := internal.HealthCheck(db, &serversList, settingsObj)

	req := httptest.NewRequest("GET", "/__internal/health_check", nil)
	req.Header.Set("Authorization", "Bearer testtoken")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got %q", w.Body.String())
	}
}

func TestHealthCheck_Unauthorized(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	serversList := []servers.Server{}
	settingsObj := &settings.Settings{
		Token: "testtoken",
	}
	handler := internal.HealthCheck(db, &serversList, settingsObj)

	req := httptest.NewRequest("GET", "/__internal/health_check", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", w.Code)
	}
}

func TestHealthCheck_DBUnreachable(t *testing.T) {
	// Use a closed DB to simulate unreachable DB
	db, _ := sql.Open("sqlite3", ":memory:")
	db.Close()
	serversList := []servers.Server{}
	settingsObj := &settings.Settings{
		Token: "testtoken",
	}
	handler := internal.HealthCheck(db, &serversList, settingsObj)

	req := httptest.NewRequest("GET", "/__internal/health_check", nil)
	req.Header.Set("Authorization", "Bearer testtoken")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 Internal Server Error, got %d", w.Code)
	}
}
