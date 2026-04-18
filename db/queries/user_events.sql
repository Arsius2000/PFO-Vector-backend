-- name: AddUserEvent :exec
INSERT INTO user_events (
    user_id,
    event_id
)
VALUES ($1, $2)
ON CONFLICT (user_id, event_id) DO NOTHING;

-- name: GetUserEventsByUserID :many
SELECT e.*
FROM events e
JOIN user_events ue ON ue.event_id = e.id
WHERE ue.user_id = $1
ORDER BY e.event_date DESC
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');


