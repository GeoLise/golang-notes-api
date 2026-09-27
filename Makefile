include .env

export 

migrate-up:
	goose -dir migrations postgres "${DATABASE_URL}" up

migrate-down:
	goose -dir migrations postgres "${DATABASE_URL}" down

migrate-new:
	goose -dir migrations create $(name) sql

migrate-status:
	goose -dir migrations postgres "${DATABASE_URL}" status

run:
	go run cmd/server/main.go