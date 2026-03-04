-- name: CreateUser :one
INSERT INTO users (
    full_name,
    telegram,
    gender,
    phone_number,
    telegram_id
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC;



-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;