package handler

import (
	"database/sql"
	"encoding/json"
	"errors"

	"net/http"
	"strconv"
	"strings"

	"pfo-vector/internal/middleware"
	"pfo-vector/internal/model"
	"pfo-vector/internal/repository"
	"pfo-vector/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
)
type UpdateUserRequest struct {
    FullName           *string `json:"full_name,omitempty"`
    Gender             *string `json:"gender,omitempty"`
    DirectionVector    *string `json:"direction_vector,omitempty"`
    StudyGroup         *string `json:"study_group,omitempty"`
    Rating             *int32  `json:"rating,omitempty"`
    VisitedEventsCount *int32  `json:"visited_events_count,omitempty"`
    PhoneNumber        *string `json:"phone_number,omitempty"`
    Telegram           *string `json:"telegram,omitempty"`
    AvatarURL          *string `json:"avatar_url,omitempty"`
    Role               *string `json:"role,omitempty"`
    TelegramID         *int32  `json:"telegram_id,omitempty"`
}

type UserHandler struct{
	queries *repository.Queries
	service *service.UserImportService
}

func NewUserHandler(queries *repository.Queries,service *service.UserImportService) *UserHandler{
	return &UserHandler{
		queries: queries,
		service: service,
	}
}

type CreateUserRequest struct {
	FullName           string  `json:"full_name"`
	Gender             *string `json:"gender,omitempty"`
	DirectionVector    *string `json:"direction_vector,omitempty"`
	StudyGroup         *string `json:"study_group,omitempty"`
	Rating             *int32  `json:"rating,omitempty"`             // Если nil -> БД поставит 0
	VisitedEventsCount *int32  `json:"visited_events_count,omitempty"` // Если nil -> БД поставит 0
	PhoneNumber        *string `json:"phone_number,omitempty"`
	Telegram           string  `json:"telegram"`                     // Обязательно (NOT NULL в БД)
	AvatarURL          *string `json:"avatar_url,omitempty"`
	TelegramID         *int32  `json:"telegram_id,omitempty"`        // UNIQUE в БД
	// Role не нужен в запросе, если нас устраивает дефолт 'боец'. 
	// Если нужно менять роль при создании, добавьте поле сюда.
}

// CreateUser godoc
// @Summary      Создание пользователя
// @Description  Создает нового пользователя с переданными данными
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      handler.CreateUserRequest  true  "Данные пользователя"
// @Success      201   {object}  model.UserResponse            "Пользователь успешно создан"
// @Failure      400   {string}  string                     "Неверный формат запроса или валидация не пройдена"
// @Failure      409   {string}  string                     "Пользователь с таким telegram_id уже существует"
// @Failure      500   {string}  string                     "Ошибка сервера"
// @Router       /users/add [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	// Валидация
	if strings.TrimSpace(req.FullName) == "" {
		http.Error(w, "Поле full_name обязательно", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Telegram) == "" {
		http.Error(w, "Поле telegram обязательно", http.StatusBadRequest)
		return
	}
	if req.TelegramID != nil && *req.TelegramID <= 0 {
		http.Error(w, "telegram_id должен быть положительным числом", http.StatusBadRequest)
		return
	}

	// --- ХЕЛПЕРЫ ДЛЯ PGTYPE ---

	// Для текстовых полей (VARCHAR, TEXT) -> pgtype.Text
	nullText := func(s *string) pgtype.Text {
		if s == nil {
			return pgtype.Text{Valid: false}
		}
		return pgtype.Text{String: *s, Valid: true}
	}

	// Для целочисленных полей (INT, INT4) -> pgtype.Int4
	nullInt4 := func(i *int32) pgtype.Int4 {
		if i == nil {
			return pgtype.Int4{Valid: false}
		}
		return pgtype.Int4{Int32: *i, Valid: true}
	}

	// --- СБОРКА ПАРАМЕТРОВ ---

	args := repository.CreateUserParams{
		FullName:           req.FullName,                // string
		Gender:             nullText(req.Gender),        // pgtype.Text
		DirectionVector:    nullText(req.DirectionVector),
		StudyGroup:         nullText(req.StudyGroup),
		Rating:             nullInt4(req.Rating),        // pgtype.Int4
		VisitedEventsCount: nullInt4(req.VisitedEventsCount),
		PhoneNumber:        nullText(req.PhoneNumber),
		Telegram:           req.Telegram,                // string (NOT NULL в БД)
		AvatarUrl:          nullText(req.AvatarURL),     // <--- Исправленное имя поля!
		TelegramID:         nullInt4(req.TelegramID),
	}

	// Выполнение запроса
	user, err := h.queries.CreateUser(ctx, args)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			http.Error(w, "Пользователь с таким telegram_id уже существует", http.StatusConflict)
			return
		}
		http.Error(w, "Ошибка сервера: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp := model.MapUserFromRepo(user)
	json.NewEncoder(w).Encode(resp)
}


// ImportUsers godoc
// @Summary      Импорт пользователей из Excel
// @Description  Загружает Excel-файл и импортирует пользователей
// @Tags         users
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "Excel файл (.xlsx)"
// @Success      201   {object}  model.ImportResult  "Импорт завершён"
// @Failure      400   {string}  string              "Некорректный файл или данные"
// @Failure      409   {string}  string              "Конфликт уникальных данных"
// @Failure      500   {string}  string              "Ошибка сервера"
// @Router       /users/import-users [post]
func  (h *UserHandler) ImportUsers(w http.ResponseWriter, r *http.Request){

	ctx := r.Context()


	//Проверка на правильность метода
	if r.Method!=http.MethodPost{
		http.Error(w,"Method not allowed",http.StatusMethodNotAllowed)
		return
	}

	//Ограничение файла в 10 мб
	err := r.ParseMultipartForm(10<<20)
	if err!=nil{
		http.Error(w,"Invalid multipart form",http.StatusBadRequest)
		return
	}
	file,_ ,err := r.FormFile("file")
	if err!=nil{
		http.Error(w,"file field 'file' is required",http.StatusBadRequest)
		return
	}
	defer file.Close()

	response,err := h.service.ImportFromExcel(ctx,file)
	if err!=nil{
		switch {
			case errors.Is(err, service.ErrOpenExcelReader):
				http.Error(w, "Error not open excel file", http.StatusBadRequest)
				return
			case errors.Is(err, service.ErrEmptyWorkbook):
				http.Error(w, "empty workbook", http.StatusBadRequest)
				return
			case errors.Is(err, service.ErrFailedReadRows):
				http.Error(w, "failed to read rows", http.StatusBadRequest)
				return
			case errors.Is(err, service.ErrNoDataRows):
				http.Error(w, "no data rows", http.StatusBadRequest)
				return
			case errors.Is(err, service.ErrMissingRequiredCol):
				http.Error(w, err.Error(), http.StatusBadRequest) // тут останется имя колонки
				return
			default:
				http.Error(w, "server error", http.StatusInternalServerError)
				return
			}
	}
	
	w.WriteHeader(http.StatusCreated)
	
	json.NewEncoder(w).Encode(response)

}

// UpdateUser godoc
// @Summary      Обновление пользователя
// @Description  Частично обновляет данные пользователя по ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      int                        true  "ID пользователя"
// @Param        user  body      handler.UpdateUserRequest  true  "Данные для обновления"
// @Success      200   {object}  model.UserResponse         "Пользователь успешно обновлен"
// @Failure      400   {string}  string                     "Неверный формат запроса"
// @Failure      404   {string}  string                     "Пользователь не найден"
// @Failure      409   {string}  string                     "Пользователь с таким telegram_id уже существует"
// @Failure      500   {string}  string                     "Ошибка сервера"
// @Router       /users/{id} [patch]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

	role := ctx.Value(middleware.CtxRole)
	userID := ctx.Value(middleware.CtxUserID)



	userID, ok := r.Context().Value(middleware.CtxUserID).(int32)
	if !ok {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	role, ok = r.Context().Value(middleware.CtxRole).(string)
	if !ok {
		http.Error(w, "invalid role", http.StatusUnauthorized)
		return
	}

    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }
	//Проверка роли ,ЕСли роль админ может менять всех иначе только себя
	if role != "админ" && userID != int32(id) {
    http.Error(w, "forbidden", http.StatusForbidden)
    return
	}
    var req UpdateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
        return
    }

    nullText := func(s *string) pgtype.Text {
        if s == nil {
            return pgtype.Text{Valid: false}
        }
        return pgtype.Text{String: *s, Valid: true}
    }
    nullInt4 := func(i *int32) pgtype.Int4 {
        if i == nil {
            return pgtype.Int4{Valid: false}
        }
        return pgtype.Int4{Int32: *i, Valid: true}
    }

    args := repository.UpdateUserParams{
        ID:                 int32(id),
        FullName:           nullText(req.FullName),
        Gender:             nullText(req.Gender),
        DirectionVector:    nullText(req.DirectionVector),
        StudyGroup:         nullText(req.StudyGroup),
        Rating:             nullInt4(req.Rating),
        VisitedEventsCount: nullInt4(req.VisitedEventsCount),
        PhoneNumber:        nullText(req.PhoneNumber),
        Telegram:           nullText(req.Telegram),
        AvatarUrl:          nullText(req.AvatarURL),
        Role:               nullText(req.Role),
        TelegramID:         nullInt4(req.TelegramID),
    }

    user, err := h.queries.UpdateUser(ctx, args)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }
        if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
            http.Error(w, "Пользователь с таким telegram_id уже существует", http.StatusConflict)
            return
        }
        http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
        return
    }

    resp := model.MapUserFromRepo(user)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode(resp)
}

// GetUser godoc
// @Summary      Получение пользователя
// @Description  Возвращает данные пользователя по ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID пользователя"
// @Success      200  {object}  model.UserResponse  "Данные пользователя"
// @Failure      404  {string}  string  "Пользователь не найден"
// @Router       /users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter,r *http.Request){

    idStr := chi.URLParam(r, "id")  // "1"
    
    id, err := strconv.ParseInt(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

	user,err :=h.queries.GetUser(r.Context(),int32(id))
	if err == sql.ErrNoRows{
		http.Error(w,"User not found",http.StatusNotFound)
		return
	}
	if err != nil{
		http.Error(w,"Database Error",http.StatusInternalServerError)
		return
	}


	//конвертация в response модель
	response := model.MapUserFromRepo(user)

	w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
		
}

// GetListUsersId godoc
// @Summary      Получение всех пользователей
// @Description  Возвращает данные пользователей по ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  model.UserListResponse  "Данные пользователей с пагинацией" 
// @Failure      404  {string}  string  "Ошибка получения списка пользователей"
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Размер страницы" default(20) minimum(1) maximum(100)
// @Router       /users/all [get]
func (h *UserHandler) ListUsersId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Парсинг параметров пагинации из URL (?page=1&limit=20)
	query := r.URL.Query()
	

	// Значения по умолчанию
	page := 1
	limit := 20 

	if v:= query.Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}
	if v:= query.Get("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil && l > 0 {
			// Ограничим максимальный размер страницы, чтобы не нагружать БД
			if l > 100 {
				limit = 100
			} else {
				limit = l
			}
		}
	}

	// Расчет OFFSET: (page - 1) * limit
	offset := (page - 1) * limit

	// 2. Подготовка аргументов для sqlc
	// sqlc сгенерирует типы int32 или int64 в зависимости от вашей БД. 
	// Обычно для LIMIT/OFFSET подходит int32, но проверьте сгенерированный код.
	args := repository.ListUsersIdParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	// 3. Выполнение запроса
	users, err := h.queries.ListUsersId(ctx, args)
	if err != nil {
		// Логирование ошибки
		http.Error(w, "Ошибка получения списка пользователей", http.StatusInternalServerError)
		return
	}

	response := model.UserListResponse{
    Users: model.MapUsersFromRepo(users),
    Pagination: model.Pagination{
        Page:   page,
        Limit:  limit,
        Offset: offset,
   	},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}


// GetListUsersName godoc
// @Summary      Получение всех пользователей
// @Description  Возвращает данные пользователей по Full_name
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  model.UserListResponse  "Данные пользователей с пагинацией"
// @Failure      404  {string}  string  "Ошибка получения списка пользователей"
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Размер страницы" default(20) minimum(1) maximum(100)
// @Router       /users/all/Name [get]
func (h *UserHandler) ListUsersName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Парсинг параметров пагинации из URL (?page=1&limit=20)
	query := r.URL.Query()
	
	pageStr := query.Get("page")
	limitStr := query.Get("limit")

	// Значения по умолчанию
	page := 1
	limit := 20 

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			// Ограничим максимальный размер страницы, чтобы не нагружать БД
			if l > 100 {
				limit = 100
			} else {
				limit = l
			}
		}
	}

	// Расчет OFFSET: (page - 1) * limit
	offset := (page - 1) * limit

	// 2. Подготовка аргументов для sqlc
	// sqlc сгенерирует типы int32 или int64 в зависимости от вашей БД. 
	// Обычно для LIMIT/OFFSET подходит int32, но проверьте сгенерированный код.
	args := repository.ListUsersNameParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	// 3. Выполнение запроса
	users, err := h.queries.ListUsersName(ctx, args)
	if err != nil {
		// Логирование ошибки
		http.Error(w, "Ошибка получения списка пользователей", http.StatusInternalServerError)
		return
	}

	response := model.UserListResponse{
    Users: model.MapUsersFromRepo(users),
    Pagination: model.Pagination{
        Page:   page,
        Limit:  limit,
        Offset: offset,
   	},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}


// GetListUsersRating godoc
// @Summary      Получение всех пользователей
// @Description  Возвращает данные пользователей по Rating
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  model.UserListResponse  "Данные пользователей с пагинацией"
// @Failure      404  {string}  string  "Ошибка получения списка пользователей"
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Размер страницы" default(20) minimum(1) maximum(100)
// @Router       /users/all/Rating [get]
func (h *UserHandler) ListUsersRating(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Парсинг параметров пагинации из URL (?page=1&limit=20)
	query := r.URL.Query()
	
	pageStr := query.Get("page")
	limitStr := query.Get("limit")

	// Значения по умолчанию
	page := 1
	limit := 20 

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			// Ограничим максимальный размер страницы, чтобы не нагружать БД
			if l > 100 {
				limit = 100
			} else {
				limit = l
			}
		}
	}

	// Расчет OFFSET: (page - 1) * limit
	offset := (page - 1) * limit

	// 2. Подготовка аргументов для sqlc
	// sqlc сгенерирует типы int32 или int64 в зависимости от вашей БД. 
	// Обычно для LIMIT/OFFSET подходит int32, но проверьте сгенерированный код.
	args := repository.ListUsersRatingParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	// 3. Выполнение запроса
	users, err := h.queries.ListUsersRating(ctx, args)
	if err != nil {
		// Логирование ошибки
		http.Error(w, "Ошибка получения списка пользователей", http.StatusInternalServerError)
		return
	}


	response := model.UserListResponse{
    Users: model.MapUsersFromRepo(users),
    Pagination: model.Pagination{
        Page:   page,
        Limit:  limit,
        Offset: offset,
   	},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}


// DeleteUser godoc
// @Summary      Удаление пользователя
// @Description  Удаляет пользователя по ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID пользователя"
// @Success      200  {object}  model.UserResponse  "Пользователь удален"
// @Success      204  {string}  string "Пользователь удален"
// @Failure      404  {string}  string  "Пользователь не найден"
// @Security BearerAuth 
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {

    idStr := chi.URLParam(r, "id")  // "1"
    
    id, err := strconv.ParseInt(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }
    
    err = h.queries.DeleteUser(r.Context(), int32(id))  // Выполняем DELETE
    if err == sql.ErrNoRows {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    if err != nil {
        http.Error(w, "Failed to delete user", http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusNoContent)  // 204 — успешно, но без тела ответа
}

