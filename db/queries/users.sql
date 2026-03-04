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

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC;

-- name: UpdateUser :one
UPDATE users
SET full_name=$2, gender=$3, direction_vector=$4,study_group=$5,rating=$6,visited_events_count=$7,phone_number=$8,telegram=$9,avatar_url=$10,role=$11,telegram_id=$12
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;