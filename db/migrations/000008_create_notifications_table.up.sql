--Создаем таблицу notifications
CREATE TABLE IF NOT EXISTS notifications(
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    event_id INT NOT NULL,
    notification_date TIMESTAMP,
    sent BOOLEAN ,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT fk_notification_user_id   
        FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_notification_events_id
        FOREIGN KEY (event_id) REFERENCES events(id)
);