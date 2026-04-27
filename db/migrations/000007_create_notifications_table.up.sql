-- Создаем таблицу notifications
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    event_id INT,
    notification_type VARCHAR(50) NOT NULL,
    message_text TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    queued_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,

    CONSTRAINT chk_notifications_status
        CHECK (status IN ('pending', 'queued', 'sent', 'failed')),
    CONSTRAINT chk_notifications_retry_count
        CHECK (retry_count >= 0),

    CONSTRAINT fk_notification_user_id
        FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_notification_events_id
        FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_notifications_status_created_at
    ON notifications (status, created_at);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id
    ON notifications (user_id);

CREATE INDEX IF NOT EXISTS idx_notifications_event_id
    ON notifications (event_id);