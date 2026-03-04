--создаем таблицу user_events
CREATE TABLE IF NOT EXISTS user_events(
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    event_id INT NOT NULL,
    attended BOOLEAN ,
    --связываем таблицы users и user_events
    CONSTRAINT fk_events_user_id   
        FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_events_event_id
        FOREIGN KEY (event_id) REFERENCES events(id)
);