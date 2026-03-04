-- Создаем таблицу News
CREATE TABLE IF NOT EXISTS news(
    id SERIAL PRIMARY KEY,
    title varchar(255) NOT NULL,
    short_description varchar(255) NOT NULL,
    full_description TEXT NOT NULL,
    news_date TIMESTAMP ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by INT NOT NULL,
    --связываем таблицу news и users
    CONSTRAINT fk_news_created_by   
        FOREIGN KEY (created_by) REFERENCES users(id)
);




