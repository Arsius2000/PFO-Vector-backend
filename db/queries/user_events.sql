-- name: RegisterUserWithStatus :one
WITH ev AS (
  SELECT id, event_date, participants_current, participants_limit, end_time
  FROM events
  WHERE  events.id = sqlc.arg('event_id')
  FOR UPDATE
),
attempt_insert AS (
  INSERT INTO user_events (user_id, event_id)
  SELECT sqlc.arg('user_id'), sqlc.arg('event_id')
  FROM ev
  WHERE (
      ev.event_date > CURRENT_DATE
      OR (
        ev.event_date = CURRENT_DATE
        AND COALESCE(ev.end_time, TIME '23:59:59') > LOCALTIME
      )
    )
    AND ev.participants_current < ev.participants_limit
  ON CONFLICT (user_id, event_id) DO NOTHING
  RETURNING event_id
),
attempt_update AS (
  UPDATE events
  SET participants_current = participants_current + 1
  WHERE id = (SELECT event_id FROM attempt_insert)
  RETURNING id
)
SELECT 
  CASE 
    WHEN NOT EXISTS (SELECT 1 FROM ev) THEN 'NOT_FOUND'
    WHEN EXISTS (SELECT 1 FROM attempt_update) THEN 'SUCCESS'
    WHEN NOT EXISTS (
      SELECT 1
      FROM ev
      WHERE (
        ev.event_date > CURRENT_DATE
        OR (
          ev.event_date = CURRENT_DATE
          AND COALESCE(ev.end_time, TIME '23:59:59') > LOCALTIME
        )
      )
    ) THEN 'REGISTRATION_CLOSED'
    WHEN NOT EXISTS (SELECT 1 FROM ev WHERE participants_current < participants_limit) THEN 'NO_VACANCY'
    WHEN NOT EXISTS (SELECT 1 FROM attempt_insert) THEN 'ALREADY_REGISTERED'
    ELSE 'UNKNOWN_ERROR'
  END AS status;

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


