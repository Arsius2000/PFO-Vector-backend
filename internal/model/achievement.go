package model



type AchievementResponse struct {
	ID              int32              `json:"id"`
	AchivementsName string  `json:"achivements_name"`
	IconName        string  `json:"icon_name"`
	Description     string  `json:"description"`
	ConditionType   *string `json:"condition_type"`
	ConditionValue  *int32  `json:"condition_value"`
}

type AchievementListResponse struct{
	Achievements []AchievementResponse `json:"achievements"`
	Pagination Pagination `json:"pagination"`
}