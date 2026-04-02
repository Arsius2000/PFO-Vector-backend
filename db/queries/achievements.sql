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


-- name: UpdateAchievements :one

-- name: UpdateAchievements :one
UPDATE achievements
SET achivements_name = $2,icon_name = $3,description = $4,condition_type = $5,condition_value = $6
WHERE id = $1
RETURNING *;

-- name: DeleteAchievements :exec
DELETE FROM achievements
WHERE id = $1;