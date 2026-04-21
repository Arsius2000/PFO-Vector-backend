package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"pfo-vector/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// TelegramAuthHandler обрабатывает авторизацию через Telegram бота
type TelegramAuthHandler struct {
	queries *repository.Queries
}

// NewTelegramAuthHandler создаёт новый хендлер
func NewTelegramAuthHandler(queries *repository.Queries) *TelegramAuthHandler {
	return &TelegramAuthHandler{
		queries: queries,
	}
}

// TelegramAuthRequest — данные, которые пришлёт бот
type TelegramAuthRequest struct {
	TelegramID  int64  `json:"telegram_id"`
	Username    string `json:"username"`
	PhoneNumber string `json:"phone_number"`
}

// TelegramAuthResponse — ответ сервера
type TelegramAuthResponse struct {
	Token   string `json:"token"`
	UserID  int32  `json:"user_id"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type TelegramCheckResponse struct {
	Exists  bool   `json:"exists"`
	UserID  int32  `json:"user_id,omitempty"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// TelegramAuth godoc
// @Summary      Авторизация через Telegram
// @Description  Авторизация только для существующих пользователей. Автоматически обновляет telegram username
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  handler.TelegramAuthRequest  true  "Данные от Телеграм-бота"
// @Success      200  {object}  handler.TelegramAuthResponse  "Успешная авторизация"
// @Failure      400  {string}  string  "Неверный формат запроса"
// @Failure      401  {string}  string  "Пользователь не найден"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /auth/telegram [post]
func (h *TelegramAuthHandler) TelegramAuth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Парсим входящий JSON от бота
	var req TelegramAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	// 2. Валидация: telegram_id обязателен
	// if req.TelegramID <= 0 {
	// 	http.Error(w, "telegram_id обязателен и должен быть положительным", http.StatusBadRequest)
	// 	return
	// }
	if strings.TrimSpace(req.Username) == "" {
		http.Error(w, "username обязателен", http.StatusBadRequest)
		return
	}

	// 
	var user repository.User
	var err error
	//ЕСли телеграм id не равен нулю ищем по TelegramID
	if req.TelegramID!=0{
		user, err = h.queries.GetUserByTelegramID(ctx, pgtype.Int4{Int32: int32(req.TelegramID), Valid: true})
		if err != nil {
			 
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}
	
	}

	//ЕСли равен 0 ,то ищем по @USErname
	user, err = h.queries.GetUserByTelegramUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Пользователь не найден", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	
	// Закоментил на момент тестов
	// 4. Обновляем telegram-данные существующего пользователя
	// updateParams := repository.UpdateUserTelegramDataParams{
	// 	ID:          user.ID,
	// 	Telegram:    pgtype.Text{String: req.Username, Valid: true},
	// 	TelegramID:  pgtype.Int4{Int32: int32(req.TelegramID), Valid: true},
	// 	PhoneNumber: pgtype.Text{String: req.PhoneNumber, Valid: req.PhoneNumber != ""},
	// }
	// user, err = h.queries.UpdateUserTelegramData(ctx, updateParams)
	// if err != nil {
	// 	http.Error(w, "Ошибка обновления данных: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// 5. Генерируем JWT токен
	token, err := generateJWT(user.ID, user.TelegramID.Int32,user.Role.String)
	if err != nil {
		http.Error(w, "Ошибка генерации токена", http.StatusInternalServerError)
		return
	}

	// 6. Формируем ответ
	response := TelegramAuthResponse{
		Token:   token,
		UserID:  user.ID,
		Status:  "success",
		Message: "Авторизация успешна",
	}

	// 7. Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CheckTelegramUsername godoc
// @Summary      Проверка существования пользователя по telegram_username
// @Description  Проверяет, существует ли пользователь по telegram username
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        telegram_username   path      string  true  "Telegram username"
// @Success      200  {object}  handler.TelegramCheckResponse  "Результат проверки"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /auth/check/{telegram_username} [get]
func (h *TelegramAuthHandler) CheckTelegramUsername(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := strings.TrimSpace(chi.URLParam(r, "telegram_username"))
	if username == "" {
		http.Error(w, "telegram_username обязателен", http.StatusBadRequest)
		return
	}

	user, err := h.queries.GetUserByTelegramUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(TelegramCheckResponse{
				Exists:  false,
				Status:  "not_found",
				Message: "Пользователь не найден",
			})
			return
		}
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(TelegramCheckResponse{
		Exists:  true,
		UserID:  user.ID,
		Status:  "found",
		Message: "Пользователь найден",
	})
}


// generateJWT создаёт JWT токен для пользователя
func generateJWT(userID int32, telegramID int32,role string) (string, error) {
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
		"exp":         time.Now().Add(24 * time.Hour).Unix(), // Токен живёт 24 часа
		"issued_at":   time.Now().Unix(),
	}

	// Создаём и подписываем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
