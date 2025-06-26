run:
	go run ./cmd/main.go

test:
	go test ./...

migrate:
	go test -run=TestDBAndMigrations ./pkg/db 