package model

import "time"

type UserResponse struct {
    ID                 int32      `json:"id"`
    FullName           string     `json:"full_name"`
    Gender             *string    `json:"gender,omitempty"`
    DirectionVector    *string    `json:"direction_vector,omitempty"`
    StudyGroup         *string    `json:"study_group,omitempty"`
    Rating             int32      `json:"rating"`
    VisitedEventsCount int32      `json:"visited_events_count"`
    PhoneNumber        *string    `json:"phone_number,omitempty"`
    Telegram           string     `json:"telegram"`
    AvatarURL          *string    `json:"avatar_url,omitempty"`
    Role               string     `json:"role"`
    TelegramID         *int32     `json:"telegram_id,omitempty"`
    JoinDate           *time.Time `json:"join_date,omitempty"`
}

type Pagination struct {
    Page       int `json:"page"`
    Limit     int `json:"limit"`    
    Offset    int `json:"offset"`
}

type UserListResponse struct {
    Users []UserResponse `json:"users"`
    Pagination Pagination `json:"pagination"`
}

type UserImportResponse struct{
    Users []UserResponse `json:"users"`
    
}