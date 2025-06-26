package db

import (
	"database/sql"
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL
	);`,
	`CREATE TABLE IF NOT EXISTS deliveries (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id),
		item TEXT NOT NULL,
		delivered_at TIMESTAMP
	);`,
}

func RunMigrations(db *sql.DB) error {
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return err
		}
	}
	return nil
}
