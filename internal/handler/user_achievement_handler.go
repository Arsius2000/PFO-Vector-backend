package handler

import (
	"encoding/json"
	"net/http"
	"pfo-vector/internal/middleware"
	"pfo-vector/internal/repository"
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
// @Tags         UserAchievement
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


