package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"strings"
	"time"

	"pfo-vector/internal/model"
	"pfo-vector/internal/repository"
	"pfo-vector/internal/service"

	"github.com/go-chi/chi/v5"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

// TelegramAuthHandler обрабатывает авторизацию через Telegram бота
type TelegramAuthHandler struct {
	queries *repository.Queries
	rdb *redis.Client
}

// NewTelegramAuthHandler создаёт новый хендлер
func NewTelegramAuthHandler(queries *repository.Queries,rdb *redis.Client) *TelegramAuthHandler {
	return &TelegramAuthHandler{
		queries: queries,
		rdb:rdb,
	}
}

// TelegramAuthRequest — данные, которые пришлёт бот
type TelegramAuthRequest struct {
	TelegramID  int64  `json:"telegram_id"`
	Username    string `json:"username"`
	PhoneNumber string `json:"phone_number"`
}



// TelegramAuth godoc
// @Summary      Авторизация через Telegram
// @Description  Авторизация только для существующих пользователей. Автоматически обновляет telegram username
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  handler.TelegramAuthRequest  true  "Данные от Телеграм-бота"
// @Success      200  {object}  model.TelegramAuthResponse  "Успешная авторизация"
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
	code, err := service.GenerateCode()
	if err != nil {
		http.Error(w, "Ошибка генерации токена", http.StatusInternalServerError)
		return
	}

	// Сохраняем код в Redis: ключ = code, значение = userID, время жизни = 5 минут
	err = h.rdb.Set(ctx, code, user.ID, 5*time.Minute).Err()
	if err != nil {
		http.Error(w, "Ошибка сохранения кода", http.StatusInternalServerError)
		return
	}
	
	// 6. Формируем ответ
	response := model.TelegramAuthResponse{
		Token:   code,
		UserID:  user.ID,
		Status:  "success",
		Message: "Авторизация успешна",
	}

	// 7. Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CheckPhoneNumber godoc
// @Summary      Проверка существования пользователя по номеру телефона
// @Description  Проверяет, существует ли пользователь по номеру телефона
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        phone_number   path      string  true  "Phone Number"
// @Success      200  {object}  model.TelegramCheckResponse  "Результат проверки"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /auth/check/{phone_number} [get]
func (h *TelegramAuthHandler) CheckPhoneNumber(w http.ResponseWriter, r *http.Request) {
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
			_ = json.NewEncoder(w).Encode(model.TelegramCheckResponse{
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
	_ = json.NewEncoder(w).Encode(model.TelegramCheckResponse{
		Exists:  true,
		UserID:  user.ID,
		Status:  "found",
		Message: "Пользователь найден",
	})
}




// ExchangeCode godoc
// @Summary      Обмен кода на JWT (в HttpOnly Cookie)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        code path string true "Временный код"
// @Success      200 {string} string "Успешно, JWT в cookie"
// @Failure      400 {string} string "Неверный код"
// @Router       /auth/exchange-code/{code} [get]
func (h *TelegramAuthHandler) ExchangeCode(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    code := chi.URLParam(r, "code") 
    
    if code == "" {
        http.Error(w, "Код не найден в URL", http.StatusBadRequest)
        return
    }

    // 2. Ищем этот код в Redis
    userIDStr, err := h.rdb.Get(ctx, code).Result()
    if err != nil {
        http.Error(w, "Код не найден или истек", http.StatusBadRequest)
        return
    }
	userId,err := strconv.ParseInt(userIDStr,10,32)
    if err!=nil{
		http.Error(w,"Неверный формат ID",http.StatusBadRequest)
	}
    // 3. Удаляем код (одноразовый)
    h.rdb.Del(ctx, code)
    
    // 4. Получаем юзера 
	user,err := h.queries.GetUser(ctx,int32(userId))
    
    // 5. Генерируем JWT
    token, err := service.GenerateJWT(int32(userId),user.TelegramID.Int32,user.Role.String)  
	if err != nil {
		http.Error(w, "Ошибка генерации токена", http.StatusInternalServerError)
		return
	}  
    // 6. Устанавливаем HttpOnly Cookie
    // http.SetCookie(w, &http.Cookie{
    //     Name:     "jwt_token",
    //     Value:    token,
    //     Path:     "/",
    //     MaxAge:   86400, // 24 часа
    //     HttpOnly: true,  // <-- JS не может прочитать!
    //     Secure:   false, // true, если используете HTTPS
    // })


	response := model.TelegramAuthResponse{
		Token:   token,
		UserID:  user.ID,
		Status:  "success",
		Message: "Авторизация успешна",
	}

	// 7. Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
    
    // w.WriteHeader(http.StatusOK)
    w.Write([]byte("Авторизация успешна"))
}