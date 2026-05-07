

-- Создаем таблицу users
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    gender VARCHAR(50),
    direction_vector VARCHAR(255),
    study_group VARCHAR(100),
    rating INT DEFAULT 0,
    visited_events_count INT DEFAULT 0,
    phone_number VARCHAR(20) NOT NULL,
    telegram VARCHAR(100) NOT NULL UNIQUE,
    avatar_url TEXT,
    join_date TIMESTAMPTZ DEFAULT NOW(),
    role VARCHAR(50) DEFAULT 'боец',
    telegram_id INT UNIQUE

);



