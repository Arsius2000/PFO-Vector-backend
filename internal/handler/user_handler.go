package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"pfo-vector/internal/database"

	"github.com/go-chi/chi/v5"
)


type UserHandler struct{
	quries *database.Queries
}

func NewUserHandler(queries *database.Queries) *UserHandler{
	return &UserHandler{
		quries: queries,
	}
}

//Get /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter,r *http.Request){

    idStr := chi.URLParam(r, "id")  // "1"
    
    id, err := strconv.ParseInt(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

	user,err :=h.quries.GetUser(r.Context(),int32(id))
	if err == sql.ErrNoRows{
		http.Error(w,"User not found",http.StatusNotFound)
		return
	}
	if err != nil{
		http.Error(w,"Database Error",http.StatusInternalServerError)
	}


	//конвертация в response модель
	response := database.User{
		ID:int32(user.ID),
		FullName: user.FullName,
		Gender: user.Gender,
		DirectionVector: user.DirectionVector,
		StudyGroup: user.StudyGroup,
		Rating: user.Rating,
		VisitedEventsCount: user.VisitedEventsCount,
		PhoneNumber: user.PhoneNumber,
		Telegram: user.Telegram,
		JoinDate: user.JoinDate,
		Role: user.Role,
		TelegramID: user.TelegramID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,

	}
	w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
		
}

