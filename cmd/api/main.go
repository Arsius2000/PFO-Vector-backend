package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"pfo-vector/internal/handler"
	"pfo-vector/internal/middleware"
	db "pfo-vector/internal/repository"
	"pfo-vector/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title           My Project API
// @version         1.0
// @description     API сервер для управления пользователями
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @schemes         http
func main() {
	//ПОдгружаем данные из файла .env в окружение
	if err := godotenv.Load(); err != nil {
        log.Println("Warning: .env file not found, using system env vars")
    }
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
	service:=service.NewUserImportService(queries)
    userHandler := handler.NewUserHandler(queries,service)
	eventHandler := handler.NewEventHandler(queries)
	userEventHandler := handler.NewUserEventHandler(queries)
	AchievementsHandler := handler.NewAchievementsHandler(queries)
	userAchievementsHandler := handler.NewUserAchievementHandler(queries)
    
    //РУЧКИ
     r := chi.NewRouter()
		// --- Авторизация через Telegram ---
	telegramAuthHandler := handler.NewTelegramAuthHandler(queries)
	r.Post("/auth/telegram", telegramAuthHandler.TelegramAuth)
	r.Post("/auth/check/{telegram_username}", telegramAuthHandler.CheckTelegramUsername)

	//
    

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev_secret_key_change_me" // fallback для разработки
	}

	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuth(secret))

		
		r.Post("/users/import-users",userHandler.ImportUsers)
		r.Post("/profile/event/add",userEventHandler.AddUserEvent)
		r.Post("/profile/achievement/add",userAchievementsHandler.AddUserAchievement)
		r.Get("/profile/{user_id}/achievements",userAchievementsHandler.UserAchievementListId)

		r.Get("/profile/{user_id}/events" ,userEventHandler.UserEventListId)
		r.Patch("/users/{id}", userHandler.UpdateUser)
		r.Get("/users/{id}", userHandler.GetUser)
		
		r.Get("/users/all",userHandler.ListUsersId)
		
		r.Get("/profile",userHandler.GetProfile)

		r.Get("/events/{id}",eventHandler.GetEvent)
		r.Get("/events/all",eventHandler.ListEventsId)
		r.Delete("/users/{id}",userHandler.DeleteUser)
		r.Post("/users/add", userHandler.CreateUser)
		r.Post("/events/add",eventHandler.CreateEvent)

		r.Post("/achievement/add",AchievementsHandler.CreateAchievement)
		r.Get("/achievement/{id}",AchievementsHandler.GetAchievement)
		r.Get("/achievement/all",AchievementsHandler.ListAchievementsId)

	})
	// --- Подключение Swagger ---
	// Маршрут для Swagger UI будет доступен по адресу http://localhost:8080/swagger/index.html
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // URL где лежит json файл
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),

	))
	
	// Явный маршрут для JSON файла (если httpSwagger не подхватит автоматически)
	r.Get("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json")
	})

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