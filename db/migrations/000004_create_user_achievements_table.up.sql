--Создаем таблицу
CREATE TABLE  IF NOT EXISTS user_achievements(
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    achievement_id INT NOT NULL,
    awarded_at TIMESTAMPTZ DEFAULT NOW(),
    awarded_by INT NOT NULL,

    --связываем таблицу user_achievements и users
    CONSTRAINT fk_achievements_user_id   
        FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_achievements_awarded_by  
        FOREIGN KEY (awarded_by) REFERENCES users(id),
    --связываем таблицу user_achievements и achievements
    CONSTRAINT fk_achievements_achievement_id
        FOREIGN KEY (achievement_id) REFERENCES achievements(id)
 




);