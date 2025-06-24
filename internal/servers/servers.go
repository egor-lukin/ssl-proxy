package servers

import (
	"database/sql"
)

type Server struct {
	Domain  string
	SSLKey  string
	SSLCert string
}

func CreateTable(db *sql.DB) error {
	createTable := `
	CREATE TABLE IF NOT EXISTS servers (
		domain TEXT NOT NULL UNIQUE,
		ssl_key TEXT NOT NULL,
		ssl_cert TEXT NOT NULL
	);`
	_, err := db.Exec(createTable)
	return err
}

func Select(db *sql.DB) ([]Server, error) {
	rows, err := db.Query("SELECT domain, ssl_key, ssl_cert FROM servers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []Server
	for rows.Next() {
		var s Server
		if err := rows.Scan(&s.Domain, &s.SSLKey, &s.SSLCert); err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return servers, nil
}

func Insert(db *sql.DB, s *Server) error {
	_, err := db.Exec(
		"INSERT INTO servers (domain, ssl_key, ssl_cert) VALUES (?, ?, ?)",
		s.Domain, s.SSLKey, s.SSLCert,
	)
	return err
}

func Update(db *sql.DB, s *Server) error {
	_, err := db.Exec(
		"UPDATE servers SET ssl_key = ?, ssl_cert = ? WHERE domain = ?",
		s.SSLKey, s.SSLCert, s.Domain,
	)
	return err
}
