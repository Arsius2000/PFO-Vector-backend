package handler

import (
	"encoding/json"
	"net/http"
	"pfo-vector/internal/middleware"
	"pfo-vector/internal/model"
	"pfo-vector/internal/repository"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UserAchievementHandler struct {
	queries *repository.Queries
}

func NewUserAchievementHandler(queries *repository.Queries) *UserAchievementHandler {
	return &UserAchievementHandler{
		queries: queries,
	}
}



type AddUserAchievementRequest struct{
	UserId int `json:"user_id"`
	AchievementId int `json:"achievement_id"`
}

// AddUserAchievement godoc
// @Summary      Добавление достижения пользователю
// @Description  Добавляет достижения пользователю
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        user  body      handler.AddUserAchievementRequest  true  "Данные о пользователе и достижении"
// @Success      201   {string}  string            "Достижение присвоенно"
// @Failure      500   {string}  string                     "Ошибка сервера"
// @Security BearerAuth 
// @Router       /profile/achievement/add [post]
func (h *UserAchievementHandler) AddUserAchievement(w http.ResponseWriter,r *http.Request){
	ctx :=r.Context()


	//Получение и Провера ID 
	awardedBy, ok := r.Context().Value(middleware.CtxUserID).(int32)
	if !ok {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}
	var req AddUserAchievementRequest

	if err := json.NewDecoder(r.Body).Decode(&req);err!=nil{
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	args:=repository.AddUserAchievementsParams{
		UserID: int32(req.UserId),
		AchievementID: int32(req.AchievementId),
		AwardedBy: awardedBy,
	}

	err := h.queries.AddUserAchievements(ctx,args)
	if err!=nil{
		
		http.Error(w,"Ошибка сервера: "+err.Error(),http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// UserAchievementListId godoc
// @Summary      Получение всех достижений пользователя
// @Description  Возвращает список достижений по ID
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        user_id  path   int  true  "ID пользователя"
// @Success      200  {object}  model.UserAchievementListResponse  "Данные мероприятий пользователя с пагинацией" 
// @Failure      400  {string}  string  "Некорректный user_id"
// @Failure      500  {string}  string  "Ошибка получения списка достижений"
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param limit query int false "Размер страницы" default(20) minimum(1) maximum(100)
// @Security BearerAuth 
// @Router       /profile/{user_id}/achievements [get]
func (h *UserAchievementHandler) UserAchievementListId(w http.ResponseWriter, r *http.Request) {
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
	args := repository.GetUserAchievementsByUserIDParams{
		UserID: int32(userID),
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	// 3. Выполнение запроса
    achievements, err := h.queries.GetUserAchievementsByUserID(ctx,args)
    if err != nil {
        http.Error(w, "Ошибка получения списка достижений", http.StatusInternalServerError)
        return
    }

	response := model.UserAchievementListResponse{
		UserId: int(userID),
    Achievements: model.MapAchievementsFromRepo(achievements),
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