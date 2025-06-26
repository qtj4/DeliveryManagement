package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

const dsn = "host=localhost port=5432 user=postgres password=1111 dbname=DeliveryManagement sslmode=disable"

func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
