-- name: CreateEvent :one
INSERT INTO events (
    event_date,
    start_time,
    end_time,
    title,
    audience,
    weight,
    created_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetEvent :one
SELECT * FROM events
WHERE id = $1;

-- name: ListEventsId :many
SELECT *
FROM events
ORDER BY id ASC 
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: ListEventsDate :many
SELECT *
FROM events
ORDER BY event_date ASC 
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: ListEventsTitle :many
SELECT *
FROM events
ORDER BY title ASC 
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: ListEventsByFilter :many
SELECT *
FROM events
WHERE
  sqlc.arg('filter')::text = 'all'
  OR (
    sqlc.arg('filter')::text = 'past'
    AND (
      event_date < CURRENT_DATE
      OR (event_date = CURRENT_DATE AND COALESCE(end_time, TIME '23:59:59') < LOCALTIME)
    )
  )
  OR (
    sqlc.arg('filter')::text = 'ongoing'
    AND (
      event_date = CURRENT_DATE
      AND COALESCE(start_time, TIME '00:00:00') <= LOCALTIME
      AND COALESCE(end_time, TIME '23:59:59') >= LOCALTIME
    )
  )
  OR (
    sqlc.arg('filter')::text = 'upcoming'
    AND (
      event_date > CURRENT_DATE
      OR (event_date = CURRENT_DATE AND COALESCE(start_time, TIME '00:00:00') > LOCALTIME)
    )
  )
ORDER BY event_date ASC, start_time ASC
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: UpdateEvent :one
UPDATE events
SET 
    event_date = COALESCE(sqlc.narg('event_date'), event_date),
    start_time = COALESCE(sqlc.narg('start_time'), start_time),
    end_time = COALESCE(sqlc.narg('end_time'), end_time),
    title = COALESCE(sqlc.narg('title'), title),
    audience = COALESCE(sqlc.narg('audience'), audience),
    weight = COALESCE(sqlc.narg('weight'), weight)
WHERE id = $1
RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM events
WHERE id = $1;

-- name: GetEventsByUser :many
SELECT e.* FROM events e
WHERE e.created_by = $1
ORDER BY e.event_date DESC
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');