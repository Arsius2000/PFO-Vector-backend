-- name: CreateNotification :one
INSERT INTO notifications (
    user_id,
    event_id,
    notification_type,
    message_text,
    status
)
VALUES ($1, $2, $3, $4, 'pending')
RETURNING *;

-- name: GetPendingNotificationsBatch :many
SELECT *
FROM notifications
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT sqlc.arg('limit');

-- name: MarkNotificationQueued :one
UPDATE notifications
SET
    status = 'queued',
    queued_at = NOW(),
    last_error = NULL
WHERE id = $1
  AND status = 'pending'
RETURNING *;

-- name: MarkNotificationSent :one
UPDATE notifications
SET
    status = 'sent',
    sent_at = NOW(),
    last_error = NULL
WHERE id = $1
  AND status IN ('queued', 'pending')
RETURNING *;

-- name: MarkNotificationFailed :one
UPDATE notifications
SET
    status = 'failed',
    retry_count = retry_count + 1,
    last_error = $2
WHERE id = $1
RETURNING *;

-- name: RetryFailedNotification :one
UPDATE notifications
SET
    status = 'pending',
    last_error = NULL
WHERE id = $1
  AND status = 'failed'
RETURNING *;