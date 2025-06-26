package db

import (
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestDBAndMigrations(t *testing.T) {
	db, err := OpenDB()
	assert.NoError(t, err)
	defer db.Close()

	err = RunMigrations(db)
	assert.NoError(t, err)

	// Check users table
	var exists bool
	err = db.QueryRow(`SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'users'
	)`).Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists, "users table should exist")

	// Check deliveries table
	err = db.QueryRow(`SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'deliveries'
	)`).Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists, "deliveries table should exist")
}
