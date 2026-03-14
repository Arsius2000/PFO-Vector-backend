-- name: CreateUser :one
INSERT INTO users (
    full_name,
    gender,
    direction_vector,
    study_group,
    rating,
    visited_events_count,
    phone_number,
    telegram,
    avatar_url,
    telegram_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: ListUsersId :many
SELECT * FROM users
ORDER BY id DESC;

-- name: ListUsersName :many
SELECT * FROM users
ORDER BY full_name DESC;

-- name: ListUsersRating :many
SELECT * FROM users
ORDER BY rating DESC;

-- name: UpdateUser :one
UPDATE users
SET 
    full_name = COALESCE(sqlc.narg('full_name'), full_name),
    gender = COALESCE(sqlc.narg('gender'), gender),
    direction_vector = COALESCE(sqlc.narg('direction_vector'), direction_vector),
    study_group = COALESCE(sqlc.narg('study_group'), study_group),
    rating = COALESCE(sqlc.narg('rating'), rating),
    visited_events_count = COALESCE(sqlc.narg('visited_events_count'), visited_events_count),
    phone_number = COALESCE(sqlc.narg('phone_number'), phone_number),
    telegram = COALESCE(sqlc.narg('telegram'), telegram),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    role = COALESCE(sqlc.narg('role'), role),
    telegram_id = COALESCE(sqlc.narg('telegram_id'), telegram_id)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;