package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"pfo-vector/internal/model"
	"pfo-vector/internal/repository"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AchievementHandler struct {
	queries *repository.Queries
}

func NewAchievementsHandler(queries *repository.Queries) *AchievementHandler {
	return &AchievementHandler{
		queries: queries,
	}
}

type CreateAchievementRequest struct {
	AchivementsName string  `json:"achivements_name"`
	IconName        string  `json:"icon_name"`
	Description     string  `json:"description"`
	ConditionType   *string `json:"condition_type"`
	ConditionValue  *int32  `json:"condition_value"`
}

// CreateAchievements godoc
// @Summary      Создание достижения
// @Description  Создает нового достижения с переданными данными
// @Tags         achievement
// @Accept       json
// @Produce      json
// @Param        user  body      handler.CreateAchievementRequest  true  "Данные достижения"
// @Success      201   {object}  model.AchievementResponse            "Достижение успешно создан"
// @Failure      400   {string}  string                     "Неверный формат запроса или валидация не пройдена"
// @Failure      500   {string}  string                     "Ошибка сервера"
// @Security BearerAuth 
// @Router       /achievement/add [post]
func (h *AchievementHandler) CreateAchievement(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	var req CreateAchievementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	//Валидация
	if strings.TrimSpace(req.AchivementsName) == "" {
		http.Error(w, "Поле achivements_name обязательно", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.IconName) == "" {
		http.Error(w, "Поле icon_name обязательно", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Description) == "" {
		http.Error(w, "Поле description обязательно", http.StatusBadRequest)
		return
	}

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
	args := repository.CreateAchievementsParams{
		AchivementsName: req.AchivementsName,
		IconName:        req.IconName,
		Description:     req.Description,
		ConditionType:   nullText(req.ConditionType),
		ConditionValue:  nullInt4(req.ConditionValue),
	}
	achivement,err := h.queries.CreateAchievements(ctx,args)
	if err!=nil{
		http.Error(w, "Ошибка сервера: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	resp := model.MapAchievementFromRepo(achivement)
	json.NewEncoder(w).Encode(resp)

}



// ListAchievementsId godoc
// @Summary      Получение всех достижений
// @Description  Возвращает данные достижений по ID
// @Tags         achievement
// @Accept       json
// @Produce      json
// @Success      200  {object}  model.UserListResponse  "Данные пользователей с пагинацией" 
// @Failure      404  {string}  string  "Ошибка получения списка пользователей"
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Размер страницы" default(20) minimum(1) maximum(100)
// @Security BearerAuth 
// @Router       /achievement/all [get]
func (h *AchievementHandler) ListAchievementsId(w http.ResponseWriter, r *http.Request) {
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
	args := repository.ListAchievementsIdParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	// 3. Выполнение запроса
	achivements, err := h.queries.ListAchievementsId(ctx, args)
	if err != nil {
		// Логирование ошибки
		http.Error(w, "Ошибка получения списка Достижений", http.StatusInternalServerError)
		return
	}

	response := model.AchievementListResponse{
    Achievements: model.MapAchievementsFromRepo(achivements),
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



// GetAchievement godoc
// @Summary      Получение достижения
// @Description  Возвращает данные достижения по ID
// @Tags         achievement
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID достижения"
// @Success      200  {object}  model.AchievementResponse  "Данные достижения"
// @Failure      404  {string}  string  "Достижение не найдено"
// @Security BearerAuth 
// @Router       /achievement/{id} [get]
func (h *AchievementHandler) GetAchievement(w http.ResponseWriter,r *http.Request){

    idStr := chi.URLParam(r, "id")  // "1"

	
    
    id, err := strconv.ParseInt(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

	achievement,err :=h.queries.GetAchievement(r.Context(),int32(id))
	if errors.Is(err,pgx.ErrNoRows){
		http.Error(w,"User not found",http.StatusNotFound)
		return
	}
	if err != nil{
		http.Error(w,"Database Error",http.StatusInternalServerError)
		return
	}


	//конвертация в response модель
	response := model.MapAchievementFromRepo(achievement)

	w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
		
}