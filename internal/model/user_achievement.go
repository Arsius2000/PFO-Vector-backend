package model

type UserAchievementResponse struct{
	UserId int `json:"user_id"`
	AchievementId int `json:"achievement_id"`
}


type UserAchievementListResponse struct{
	UserId int `json:"user_id"`
	Achievements []AchievementResponse `json:"achievement_list_response"`
    Pagination Pagination `json:"pagination"`
}
