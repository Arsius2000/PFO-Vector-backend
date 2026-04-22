-- name: CreateAchievements :one
INSERT INTO achievements(
    achivements_name,
    icon_name,
    description,
    condition_type,
    condition_value
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAchievement :one
SELECT * FROM achievements
WHERE id = $1;

-- name: ListAchievementsId :many
SELECT *
FROM achievements
ORDER BY id ASC 
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');


-- name: DeleteAchievements :exec
DELETE FROM achievements
WHERE id = $1;