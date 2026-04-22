--Создаем таблицу events
CREATE TABLE IF NOT EXISTS events(
    id SERIAL PRIMARY KEY,
    event_date DATE NOT NULL,
    start_time TIME,
    end_time TIME,
    title TEXT,
    audience VARCHAR(150),
    weight INT,
    
    created_by INT NOT NULL,
    --Связываем таблицу users и events
    CONSTRAINT fk_events_user_id   
        FOREIGN KEY (created_by) REFERENCES users(id)

);