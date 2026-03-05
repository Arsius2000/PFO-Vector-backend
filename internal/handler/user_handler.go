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
	queries *database.Queries
}

func NewUserHandler(queries *database.Queries) *UserHandler{
	return &UserHandler{
		queries: queries,
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

	user,err :=h.queries.GetUser(r.Context(),int32(id))
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

//DELETE /users/{id}
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

