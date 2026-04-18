package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"pfo-vector/internal/model"
	"pfo-vector/internal/repository"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type EventHandler struct {
	queries *repository.Queries
}

func NewEventHandler(queries *repository.Queries) *EventHandler {
	return &EventHandler{
		queries: queries,
	}
}

type CreateEventRequest struct {
	EventDate model.CustomDate  `json:"event_date"`
	StartTime *model.CustomTime `json:"start_time"`
	EndTime   *model.CustomTime `json:"end_time"`
	Title     *string           `json:"title"`
	Audience  *string           `json:"audience"`
	Weight    *int32            `json:"weight"`
	CreatedBy int32             `json:"created_by"`
}

// CreateEvent godoc
// @Summary      Создание мероприятия
// @Description  Создает новое мероприятие с переданными данными
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        event  body      handler.CreateEventRequest  true  "Данные мероприятия"
// @Success      201   {object}  model.EventResponse            "Мероприятие успешно создано"
// @Failure      400   {string}  string                     "Неверный формат запроса или валидация не пройдена"
// @Failure      500   {string}  string                     "Ошибка сервера"
// @Router       /events/add [post]
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
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

	// Конвертация *time.Time → pgtype.Time
	nullTime := func(t *time.Time) pgtype.Time {
		pt := pgtype.Time{}
		if t != nil {
			pt.Scan(*t) // или используй Set
		}
		return pt
	}
	// Конвертация time.Time → pgtype.Date
	nullDate := func(t time.Time) pgtype.Date {
		return pgtype.Date{Time: t, Valid: true}
	}

	args := repository.CreateEventParams{
		EventDate: nullDate(time.Time(req.EventDate)),
		StartTime: nullTime((*time.Time)(req.StartTime)),
		EndTime:   nullTime((*time.Time)(req.EndTime)),
		Title:     nullText(req.Title),
		Audience:  nullText(req.Audience),
		Weight:    nullInt4(req.Weight),
		CreatedBy: req.CreatedBy,
	}

	event, err := h.queries.CreateEvent(ctx, args)
	if err != nil {

		http.Error(w, "Ошибка Сервера: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := model.MapEventFromRepo(event)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)

}

// GetEvent godoc
// @Summary      Получение мероприятия
// @Description  Возвращает данные мероприятия по ID
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID мероприятия"
// @Success      200  {object}  model.EventResponse  "Данные мероприятия"
// @Failure      404  {string}  string  "Мероприятие не найдено"
// @Router       /events/{id} [get]
func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id") // "1"

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	event, err := h.queries.GetEvent(r.Context(), int32(id))
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}

	//конвертация в response модель
	response := model.MapEventFromRepo(event)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

// GetListEventsId godoc
// @Summary      Получение всех мероприятий
// @Description  Возвращает данные мероприятий по ID
// @Tags         events
// @Accept       json
// @Produce      json
// @Success      200  {object}  model.EventsListResponse  "Данные   с пагинацией"
// @Failure      404  {string}  string  "Ошибка получения списка мероприятий"
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Размер страницы" default(20) minimum(1) maximum(100)
// @Router       /events/all [get]
func (h *EventHandler) ListEventsId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Парсинг параметров пагинации из URL (?page=1&limit=20)
	query := r.URL.Query()

	// Значения по умолчанию
	page := 1
	limit := 20

	if v := query.Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}
	if v := query.Get("limit"); v != "" {
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
	args := repository.ListEventsIdParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	// 3. Выполнение запроса
	events, err := h.queries.ListEventsId(ctx, args)
	if err != nil {
		// Логирование ошибки
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
