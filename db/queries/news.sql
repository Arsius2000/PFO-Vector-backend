-- name: CreateNews :one
INSERT INTO news(
    title,
    short_description,
    full_description,
    news_date,
    created_by
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetNews :one
SELECT * FROM news
WHERE id = $1;

-- name: ListNews :many
SELECT * FROM news
ORDER BY created_at DESC;


-- name: UpdateNews :one
UPDATE news
SET title = $2, short_description=$3,full_description=$4,news_date=$5
WHERE id  = $1
RETURNING *;

-- name: DeleteNews :exec
DELETE FROM news
WHERE id = $1;
