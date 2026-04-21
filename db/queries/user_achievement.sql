-- name: AddUserAchievements :exec
INSERT INTO user_achievements (
    user_id,
    achievement_id,
    awarded_by
)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, achievement_id) DO NOTHING;

-- name: GetUserAchievementsByUserID :many
SELECT a.*
FROM achievements a
JOIN user_achievements ua ON ua.achievement_id = a.id
WHERE ua.user_id = $1
ORDER BY ua.awarded_at DESC
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');


