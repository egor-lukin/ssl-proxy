package internal

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"

	"github.com/egor-lukin/ssl-proxy/internal/servers"
	"github.com/egor-lukin/ssl-proxy/internal/settings"
)

type SslCertsGenerator interface {
	CreateCerts(domain, email string) (string, string, error)
}

func PrepareProxy(dbPath, proxyDomain, destinationDomain, email string, sslGenerator SslCertsGenerator) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	err = settings.CreateTable(db)
	if err != nil {
		return nil, err
	}
	log.Println("Settings table prepared.")

	err = servers.CreateTable(db)
	if err != nil {
		return nil, err
	}
	log.Println("Servers table prepared.")

	settingsList, err := settings.Select(db)
	var settingsObj settings.Settings
	if len(settingsList) == 0 {
		log.Println("Existing settings do not found")

		settingsObj = settings.Settings{
			ProxyDomain:       proxyDomain,
			DestinationDomain: destinationDomain,
			Email:             email,
		}
		err = settings.Insert(db, &settingsObj)
		if err != nil {
			return nil, err
		}
		// After insert, fetch the settings again
		settingsList, err = settings.Select(db)
		if err != nil {
			return nil, err
		}
	}
	if len(settingsList) > 0 {
		settingsObj = settingsList[0]
	}

	serversList, err := servers.Select(db)
	if err != nil {
		return nil, err
	}
	log.Printf("Found %d servers in the database.", len(serversList))

	if len(serversList) == 0 {
		privateKey, certificate, err := sslGenerator.CreateCerts(proxyDomain, email)
		if err != nil {
			return nil, err
		}

		serverObj := &servers.Server{
			Domain:  proxyDomain,
			SSLKey:  privateKey,
			SSLCert: certificate,
		}
		err = servers.Insert(db, serverObj)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("Settings: %+v", settingsObj)

	return db, nil
}
