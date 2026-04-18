-- name: GetUserByTelegramID :one
SELECT * FROM users
WHERE telegram_id = sqlc.arg('telegram_id')
LIMIT 1;

-- name: GetUserByTelegramUsername :one
SELECT * FROM users
WHERE telegram = sqlc.arg('telegram')
LIMIT 1;

-- name: UpdateUserTelegramData :one
UPDATE users
SET 
    telegram = COALESCE(sqlc.narg('telegram'), telegram),
    telegram_id = COALESCE(sqlc.narg('telegram_id'), telegram_id),
    phone_number = COALESCE(sqlc.narg('phone_number'), phone_number)
WHERE id = sqlc.arg('id')
RETURNING *;