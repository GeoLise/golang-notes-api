package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"notes-api/internal/auth"
	"notes-api/internal/bot"
	"notes-api/internal/handlers"
	"notes-api/internal/middlewares"
	"notes-api/internal/models"
	"notes-api/internal/storage"
	"os"
	"time"

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

	sessionRepo := storage.NewSessions(pool)

	createTestSession(context, sessionRepo)

	logincodesRepo := storage.NewLoginCodes(rdb)
	authHandler := handlers.NewAuth(sessionRepo, logincodesRepo)
	authHandler.Register(mux)

	authMiddlaware := middlewares.NewAuthMiddleWare(sessionRepo)

	notesRepo := storage.NewNotes(pool, rdb)
	notesHandler := handlers.NewNotes(notesRepo, authMiddlaware)
	notesHandler.Register(mux)

	port := os.Getenv("PORT")

	go func() {
		if err := bot.Run(context, os.Getenv("TELEGRAM_BOT_TOKEN"), logincodesRepo, userRepo); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Server is running on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, mux))

}

func createTestSession(ctx context.Context, sessionsStorage *storage.Sessions) {
	token := auth.NewToken()

	fmt.Println(token)

	sessionsStorage.Create(ctx, models.Session{
		TokenHash: auth.HashToken(token),
		UserId:    3,
		ExpiresAt: time.Now().Add(time.Hour * 24),
	})
}
