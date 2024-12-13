package models

import (
	"database/sql"
)

func GetLatestIP(db *sql.DB) (string, error) {
	var ip string
	err := db.QueryRow("SELECT ip FROM system_info ORDER BY id DESC LIMIT 1").Scan(&ip)
	if err != nil {
		return "", err
	}
	return ip, nil
}
