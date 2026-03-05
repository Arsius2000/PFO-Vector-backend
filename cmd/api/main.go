package main

import (
	"context"
	"log"
	"net/http"
	"os"

	db "pfo-vector/internal/database"
	"pfo-vector/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
    dbURL := os.Getenv("DATABASE_URL")
    
    // Применяем миграции перед запуском
    if err := runMigrations(dbURL); err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
    
    // Запускаем сервер
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	queries := db.New(pool)
    userHandler := handler.NewUserHandler(queries)
	newUser, err := queries.CreateUser(ctx, db.CreateUserParams{
		FullName: "Иван Иванов",
		Telegram : "@arsius2902",
		      
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Created: %+v\n", newUser)
    
    
    r := chi.NewRouter()
    
    r.Get("/users/{id}", userHandler.GetUser)
    
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
   
}
func runMigrations(dbURL string) error {
    m, err := migrate.New("file://db/migrations", dbURL)
    if err != nil {
        return err
    }
    defer m.Close()
    
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    return nil
}