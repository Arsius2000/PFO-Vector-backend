package handler

import (
	"encoding/json"
	"net/http"
	"pfo-vector/internal/model"
	"pfo-vector/internal/repository"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UserEventHandler struct {
	queries *repository.Queries
}

func NewUserEventHandler(queries *repository.Queries) *UserEventHandler{
	return &UserEventHandler{
		queries: queries,
	}
}

type AddUserEventRequest struct{
	UserId int `json:"user_id"`
    EventId int `json:"event_id"`
}

// AddUserEvent godoc
// @Summary      Добавление мероприятия пользователю
// @Description  Добавляет мероприятие пользователю
// @Tags         UserEvent
// @Accept       json
// @Produce      json
// @Param        user  body      handler.AddUserEventRequest  true  "Данные о пользователе и мероприятии"
// @Success      201   {string}  string            "Записан на мероприятие"
// @Failure      500   {string}  string                     "Ошибка сервера"
// @Router       /profile/event/add [post]
func (h *UserEventHandler) AddUserEvent(w http.ResponseWriter,r *http.Request){
	ctx:=r.Context()

	var req AddUserEventRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	args := repository.AddUserEventParams{
		UserID: int32(req.UserId),
		EventID: int32(req.EventId),
	}

	err := h.queries.AddUserEvent(ctx,args)
	if err!=nil{
				http.Error(w, "Ошибка сервера: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}


// UserEventListId godoc
// @Summary      Получение всех мероприятий пользователя
// @Description  Возвращает список мероприятий по ID
// @Tags         UserEvent
// @Accept       json
// @Produce      json
// @Param        user_id  path   int  true  "ID пользователя"
// @Success      200  {object}  model.EventsListResponse  "Данные мероприятий пользователя с пагинацией" 
// @Failure      400  {string}  string  "Некорректный user_id"
// @Failure      500  {string}  string  "Ошибка получения списка мероприятий"
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Размер страницы" default(20) minimum(1) maximum(100)
// @Router       /profile/{user_id}/events [get]
func (h *UserEventHandler) UserEventListId(w http.ResponseWriter, r *http.Request) {
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


	userIDStr := chi.URLParam(r, "user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
    http.Error(w, "Invalid user_id", http.StatusBadRequest)
    return
}
	// 2. Подготовка аргументов для sqlc
	// sqlc сгенерирует типы int32 или int64 в зависимости от вашей БД. 
	// Обычно для LIMIT/OFFSET подходит int32, но проверьте сгенерированный код.
	args := repository.GetUserEventsByUserIDParams{
		UserID: int32(userID),
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	// 3. Выполнение запроса
    events, err := h.queries.GetUserEventsByUserID(ctx,args)
    if err != nil {
        http.Error(w, "Ошибка получения списка мероприятий", http.StatusInternalServerError)
        return
    }

	response := model.EventsListResponse{
    Events: model.MapEventsFromRepo(events),
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