package internal

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/egor-lukin/ssl-proxy/internal/servers"
	"github.com/egor-lukin/ssl-proxy/internal/settings"
)

func NewServer(db *sql.DB) *http.Server {
	settingsList, err := settings.Select(db)
	if err != nil {
		log.Printf("Failed to load settings: %v", err)
	} else if len(settingsList) > 1 {
		log.Printf("SettingsList len more than 1: %v", err)
	} else {
		log.Printf("Loaded settings: %+v", settingsList)
	}

	serversList, err := servers.Select(db)
	if err != nil {
		log.Printf("Failed to load servers: %v", err)
	} else {
		log.Printf("Loaded servers: %+v", serversList)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /__internal/servers", InsertServer(db, &serversList, &settingsList[0]))
	mux.HandleFunc("PUT /__internal/servers", UpdateServer(db, &serversList, &settingsList[0]))
	mux.HandleFunc("/__internal/health_check", HealthCheck(db, &serversList, &settingsList[0]))
	mux.HandleFunc("/", ProxyHandler(db, serversList, &settingsList[0]))

	return &http.Server{
		Addr:    ":443",
		Handler: mux,
		TLSConfig: &tls.Config{
			GetCertificate: getCertificateFunc(&serversList),
		},
	}
}

func UpdateServer(db *sql.DB, serversList *[]servers.Server, settingsObj *settings.Settings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || authHeader != "Bearer "+settingsObj.Token {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var s servers.Server
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if s.Domain == "" || s.SSLKey == "" || s.SSLCert == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}
		err := servers.Update(db, &s)
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		newServersList, err := servers.Select(db)
		if err != nil {
			log.Printf("Failed to refresh servers list: %v", err)
		}
		*serversList = newServersList

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Server stored"))
	}
}

func InsertServer(db *sql.DB, serversList *[]servers.Server, settingsObj *settings.Settings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || authHeader != "Bearer "+settingsObj.Token {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var s servers.Server
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if s.Domain == "" || s.SSLKey == "" || s.SSLCert == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		err := servers.Insert(db, &s)
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		newServersList, err := servers.Select(db)
		if err != nil {
			log.Printf("Failed to refresh servers list: %v", err)
		}

		*serversList = newServersList

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Server created"))
	}
}

func ProxyHandler(db *sql.DB, serversList []servers.Server, settingsObj *settings.Settings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		found := false
		for _, srv := range serversList {
			if srv.Domain == r.Host {
				found = true
				break
			}
		}
		if !found {
			http.NotFound(w, r)
			return
		}

		targetURL := &url.URL{
			Scheme: "https",
			Host:   settingsObj.DestinationDomain,
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.ServeHTTP(w, r)
	}
}

func getCertificateFunc(serversList *[]servers.Server) func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(helloInfo *tls.ClientHelloInfo) (*tls.Certificate, error) {
		for _, srv := range *serversList {
			if srv.Domain == helloInfo.ServerName {
				cert, err := tls.X509KeyPair([]byte(srv.SSLCert), []byte(srv.SSLKey))
				if err != nil {
					return nil, err
				}
				return &cert, nil
			}
		}
		return nil, fmt.Errorf("no certificate found for domain: %s", helloInfo.ServerName)
	}
}

func HealthCheck(db *sql.DB, serversList *[]servers.Server, settingsObj *settings.Settings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || authHeader != "Bearer "+settingsObj.Token {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if err := db.Ping(); err != nil {
			http.Error(w, "Database unreachable", http.StatusInternalServerError)
			return
		}

		if settingsObj == nil {
			http.Error(w, "Settings not loaded", http.StatusInternalServerError)
			return
		}

		if serversList == nil {
			http.Error(w, "Servers not loaded", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
