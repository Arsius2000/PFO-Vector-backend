package service

import (
	"crypto/rand"
	"math/big"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// generateCode создает строку из 10 случайных символов (цифры + буквы)
func GenerateCode() (string, error) {
    // Алфавит: цифры и латинские буквы
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    length := 10
    result := make([]byte, length)
    
    for i := range result {
        // Генерируем случайное число от 0 до len(charset)-1
        num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
        if err != nil {
            return "", err
        }
        result[i] = charset[num.Int64()]
    }
    
    return string(result), nil
}



// generateJWT создаёт JWT токен для пользователя
func GenerateJWT(userID int32, telegramID int32,role string) (string, error) {
	// Берём секрет из переменных окружения
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev_secret_key_change_me" // fallback для разработки
	}

	// Создаём claims (данные внутри токена)
	claims := jwt.MapClaims{
		"user_id":     userID,
		"telegram_id": telegramID,
		"role":			role,
		"exp":         time.Now().Add(240 * time.Hour).Unix(), // Токен живёт 24 часа
		"issued_at":   time.Now().Unix(),
	}

	// Создаём и подписываем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}