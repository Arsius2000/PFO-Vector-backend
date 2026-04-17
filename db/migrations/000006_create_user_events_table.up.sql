--создаем таблицу user_events
CREATE TABLE IF NOT EXISTS user_events(
    
    user_id INT NOT NULL,
    event_id INT NOT NULL,
    attended BOOLEAN NOT NULL DEFAULT false,    
    PRIMARY KEY (user_id,event_id),
    --связываем таблицы users и user_events
    CONSTRAINT fk_events_user_id   
        FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_events_event_id
        FOREIGN KEY (event_id) REFERENCES events(id)
);