package handler

import (
	"encoding/json"
	"net/http"
	"pfo-vector/internal/repository"
	"strings"
	"pfo-vector/internal/model"

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

// CreateUser godoc
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
