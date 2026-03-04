
package main

import (
    "log"
    "os"

    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    "github.com/golang-migrate/migrate/v4"
)

func main() {
    dbURL := os.Getenv("DATABASE_URL")
    
    // Применяем миграции перед запуском
    if err := runMigrations(dbURL); err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
    
    // Запускаем сервер
   
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