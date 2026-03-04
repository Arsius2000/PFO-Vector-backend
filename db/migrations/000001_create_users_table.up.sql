

-- Создаем таблицу users
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    gender VARCHAR(50),
    direction_vector VARCHAR(255),
    study_group VARCHAR(100),
    rating INTEGER DEFAULT 0,
    visited_events_count INTEGER DEFAULT 0,
    phone_number VARCHAR(20),
    telegram VARCHAR(100) NOT NULL,
    avatar_url TEXT,
    join_date TIMESTAMPTZ DEFAULT NOW(),
    role VARCHAR(50) DEFAULT 'боец',
    telegram_id INTEGER UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Создаем индексы для ускорения поиска
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_join_date ON users(join_date);

-- Триггер для обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;  
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();