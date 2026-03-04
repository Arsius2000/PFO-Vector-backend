-- Удаляем триггер
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Удаляем функцию
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаляем индексы
DROP INDEX IF EXISTS idx_users_telegram_id;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_join_date;

-- Удаляем таблицу
DROP TABLE IF EXISTS users;