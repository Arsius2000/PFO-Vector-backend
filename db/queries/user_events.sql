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
WHERE ue.user_id = sqlc.arg('user_id')
  AND (
    sqlc.arg('filter')::text = 'all'
    OR (
      sqlc.arg('filter')::text = 'past'
      AND (
        e.event_date < CURRENT_DATE
        OR (e.event_date = CURRENT_DATE AND COALESCE(e.end_time, TIME '23:59:59') < LOCALTIME)
      )
    )
    OR (
      sqlc.arg('filter')::text = 'ongoing'
      AND (
        e.event_date = CURRENT_DATE
        AND COALESCE(e.start_time, TIME '00:00:00') <= LOCALTIME
        AND COALESCE(e.end_time, TIME '23:59:59') >= LOCALTIME
      )
    )
    OR (
      sqlc.arg('filter')::text = 'upcoming'
      AND (
        e.event_date > CURRENT_DATE
        OR (e.event_date = CURRENT_DATE AND COALESCE(e.start_time, TIME '00:00:00') > LOCALTIME)
      )
    )
  )
ORDER BY e.event_date ASC, e.start_time ASC
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');


