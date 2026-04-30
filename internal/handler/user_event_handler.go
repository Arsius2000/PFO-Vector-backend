package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"pfo-vector/internal/middleware"
	"pfo-vector/internal/model"
	"pfo-vector/internal/repository"
	"pfo-vector/internal/service"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type UserEventHandler struct {
	queries *repository.Queries
	service *service.UserEventService
}

func NewUserEventHandler(queries *repository.Queries,service *service.UserEventService) *UserEventHandler{
	return &UserEventHandler{
		queries: queries,
		service: service,
	}
}

type AddUserEventRequest struct{
    EventId int `json:"event_id"`
}

// AddUserEvent godoc
// @Summary      Добавление мероприятия пользователю
// @Description  Добавляет мероприятие пользователю
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        user  body      handler.AddUserEventRequest  true  "Данные о пользователе и мероприятии"
// @Success      201   {string}  string            "Записан на мероприятие"
// @Failure      409   {string}  string            "конфликт записей"
// @Failure      500   {string}  string                     "Ошибка сервера"
// @Security BearerAuth 
// @Router       /profile/event/add [post]
func (h *UserEventHandler) AddUserEvent(w http.ResponseWriter,r *http.Request){
	ctx:=r.Context()

	userID := ctx.Value(middleware.CtxUserID).(int32)

	var req AddUserEventRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}
	err := h.service.RegisterUserToEvent(ctx,userID,int32(req.EventId))
	if err!=nil{
		switch {
		case errors.Is(err,service.ErrEventFull):
			http.Error(w,"Event full",http.StatusConflict)
			return
		case errors.Is(err,service.ErrAlreadyRegistered):
			http.Error(w,"Already Registered",http.StatusConflict)
			return
		case errors.Is(err,service.ErrRegistrationClosed):
			http.Error(w,"Registration closed",http.StatusConflict)
			return
		case errors.Is(err,service.ErrNotFound):
			http.Error(w,"Event not found",http.StatusBadRequest)
			return
		default:
			http.Error(w,"Unknown error"+err.Error(),http.StatusInternalServerError)
			return
	}
	}
	
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Мероприятие добавлено"))
}


// UserEventListId godoc
// @Summary      Получение всех мероприятий пользователя
// @Description  Возвращает список мероприятий по ID
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        user_id  path   int  true  "ID пользователя"
// @Success      200  {object}  model.EventsListResponse  "Данные мероприятий пользователя с пагинацией" 
// @Failure      400  {string}  string  "Некорректный user_id"
// @Failure      500  {string}  string  "Ошибка получения списка мероприятий"
// @Param filter query string false "Фильтр статуса" Enums(all,past,ongoing,upcoming) default(all)
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Размер страницы" default(20) minimum(1) maximum(100)
// @Security BearerAuth 
// @Router       /profile/{user_id}/events [get]
func (h *UserEventHandler) UserEventListId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	
	// 1. Парсинг параметров пагинации из URL (?page=1&limit=20)
	query := r.URL.Query()
	
	filter := strings.ToLower(strings.TrimSpace(query.Get("filter")))
	if filter == "" {
		filter = "all"
	}

	switch filter {
	case "all", "past", "ongoing", "upcoming":
		// ok
	default:
		http.Error(w, "invalid filter value", http.StatusBadRequest)
		return
	}

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
		Filter: filter,
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