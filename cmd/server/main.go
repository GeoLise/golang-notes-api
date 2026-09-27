package main

import (
	"context"
	"log"
	"net/http"
	"notes-api/internal/handlers"
	"notes-api/internal/storage"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	context := context.Background()

	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatal(err)
	}

	rdb := redis.NewClient(opt)

	pool, err := pgxpool.New(context, os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal(err)
	}

	if err := pool.Ping(context); err != nil {
		log.Fatal("No database connection", err)
	}

	mux := http.NewServeMux()

	userRepo := storage.NewUsers(pool, rdb)
	userHandler := handlers.NewUsers(userRepo)
	userHandler.Register(mux)

	port := os.Getenv("PORT")

	log.Println("Server is running on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, mux))

}
