package model

import (
    "pfo-vector/internal/repository"
)

func MapAchievementFromRepo(u repository.Achievement) AchievementResponse {

    
	var conditionType *string
	if u.ConditionType.Valid{
		conditionType= &u.ConditionType.String
	}

	var conditionValue *int32
	if u.ConditionValue.Valid{
		conditionValue = &u.ConditionValue.Int32
	}


    return AchievementResponse{
		ID : u.ID,
		AchivementsName : u.AchivementsName,
		IconName: u.IconName,
		Description: u.Description,
		ConditionType: conditionType,
		ConditionValue: conditionValue,
	}
}

func MapAchievementsFromRepo(items []repository.Achievement) []AchievementResponse {
    out := make([]AchievementResponse, 0, len(items))
    for _, u := range items {
        out = append(out, MapAchievementFromRepo(u))
    }
    return out
}