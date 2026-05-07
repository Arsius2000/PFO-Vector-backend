package model

import (
    "time"

    "pfo-vector/internal/repository"
)

func MapUserFromRepo(u repository.User) UserResponse {
    var gender *string
    if u.Gender.Valid {
        v := u.Gender.String
        gender = &v
    }

    var directionVector *string
    if u.DirectionVector.Valid {
        v := u.DirectionVector.String
        directionVector = &v
    }

    var studyGroup *string
    if u.StudyGroup.Valid {
        v := u.StudyGroup.String
        studyGroup = &v
    }

    var avatarURL *string
    if u.AvatarUrl.Valid {
        v := u.AvatarUrl.String
        avatarURL = &v
    }

    // var telegramID *int32
    // if u.TelegramID.Valid {
    //     v := u.TelegramID.Int32
    //     telegramID = &v
    // }

    var joinDate *time.Time
    if u.JoinDate.Valid {
        t := u.JoinDate.Time
        joinDate = &t
    }

    rating := int32(0)
    if u.Rating.Valid {
        rating = u.Rating.Int32
    }

    visited := int32(0)
    if u.VisitedEventsCount.Valid {
        visited = u.VisitedEventsCount.Int32
    }

    role := ""
    if u.Role.Valid {
        role = u.Role.String
    }

    return UserResponse{
        ID:                 u.ID,
        FullName:           u.FullName,
        Gender:             gender,
        DirectionVector:    directionVector,
        StudyGroup:         studyGroup,
        Rating:             rating,
        VisitedEventsCount: visited,
        PhoneNumber:        &u.PhoneNumber,
        Telegram:           u.Telegram,
        AvatarURL:          avatarURL,
        Role:               role,
        // TelegramID:         telegramID,
        JoinDate:           joinDate,
    }
}

func MapUsersFromRepo(items []repository.User) []UserResponse {
    out := make([]UserResponse, 0, len(items))
    for _, u := range items {
        out = append(out, MapUserFromRepo(u))
    }
    return out
}