package internal_test

import (
	_ "github.com/mattn/go-sqlite3"
	"os"
	"testing"

	"github.com/egor-lukin/ssl-proxy/internal"
)

type mockSslGenerator struct{}

func (m *mockSslGenerator) CreateCerts(domain, email string) (string, string, error) {
	return "dummy-key", "dummy-cert", nil
}

func TestPrepareProxy(t *testing.T) {
	dbPath := "test_settings.db"
	defer os.Remove(dbPath)

	proxyDomain := "test-proxy.example.com"
	destinationDomain := "test-destination.example.com"
	email := "test@example.com"

	db, err := internal.PrepareProxy(dbPath, proxyDomain, destinationDomain, email, &mockSslGenerator{})
	if err != nil {
		t.Fatalf("PrepareProxy failed: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT token, proxy_domain, destination_domain, email FROM settings")
	if err != nil {
		t.Fatalf("Failed to query settings: %v", err)
	}
	defer rows.Close()

	if rows.Next() {
		var token, proxyDomainDB, destinationDomainDB, emailDB string
		if err := rows.Scan(&token, &proxyDomainDB, &destinationDomainDB, &emailDB); err != nil {
			t.Fatalf("Failed to scan settings row: %v", err)
		}
		if token == "" {
			t.Errorf("Expected non-empty token, got empty string")
		}
		if proxyDomainDB != proxyDomain {
			t.Errorf("Expected proxy_domain %q, got %q", proxyDomain, proxyDomainDB)
		}
		if destinationDomainDB != destinationDomain {
			t.Errorf("Expected destination_domain %q, got %q", destinationDomain, destinationDomainDB)
		}
		if emailDB != email {
			t.Errorf("Expected email %q, got %q", email, emailDB)
		}
	} else {
		t.Errorf("No settings row found in database")
	}

	rows, err = db.Query("SELECT domain, ssl_key, ssl_cert FROM servers")
	if err != nil {
		t.Fatalf("Failed to query servers: %v", err)
	}
	defer rows.Close()

	if rows.Next() {
		var domain, sslKey, sslCert string
		if err := rows.Scan(&domain, &sslKey, &sslCert); err != nil {
			t.Fatalf("Failed to scan server row: %v", err)
		}
		if domain != proxyDomain {
			t.Errorf("Expected domain %q, got %q", proxyDomain, domain)
		}
		if sslKey != "dummy-key" {
			t.Errorf("Expected ssl_key %q, got %q", "dummy-key", sslKey)
		}
		if sslCert != "dummy-cert" {
			t.Errorf("Expected ssl_cert %q, got %q", "dummy-cert", sslCert)
		}
	}
}
