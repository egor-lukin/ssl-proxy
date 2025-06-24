package settings

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
)

type Settings struct {
	Token             string
	ProxyDomain       string
	DestinationDomain string
	Email             string
}

func CreateTable(db *sql.DB) error {
	createSettingsTable := `
	CREATE TABLE IF NOT EXISTS settings (
		token TEXT NOT NULL,
		proxy_domain TEXT NOT NULL,
		destination_domain TEXT NOT NULL,
		email TEXT NOT NULL
	);`
	_, err := db.Exec(createSettingsTable)
	return err
}

func Select(db *sql.DB) ([]Settings, error) {
	rows, err := db.Query("SELECT token, proxy_domain, destination_domain, email FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settingsList []Settings
	for rows.Next() {
		var s Settings
		if err := rows.Scan(&s.Token, &s.ProxyDomain, &s.DestinationDomain, &s.Email); err != nil {
			return nil, err
		}
		settingsList = append(settingsList, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return settingsList, nil
}

func Insert(db *sql.DB, s *Settings) error {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return err
	}
	s.Token = hex.EncodeToString(b)

	_, err = db.Exec(
		"INSERT INTO settings (token, proxy_domain, destination_domain, email) VALUES (?, ?, ?, ?)",
		s.Token, s.ProxyDomain, s.DestinationDomain, s.Email,
	)
	return err
}
