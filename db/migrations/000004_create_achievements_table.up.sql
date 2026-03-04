--Создаем таблицу 
CREATE TABLE IF NOT EXISTS achievements(
    id SERIAL PRIMARY KEY,
    achivements_name VARCHAR(255) NOT NULL ,
    icon_name VARCHAR(255) NOT NULL ,
    description TEXT NOT NULL ,
    condition_type VARCHAR(255) ,
    condition_value INT ,
    created_at TIMESTAMPTZ DEFAULT NOW()


);

