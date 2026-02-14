-- Migration: Create configs table
-- Description: Создает таблицу для хранения конфигураций различных окружений
-- Run: Автоматически при первом запуске PostgreSQL через docker-compose

CREATE TABLE IF NOT EXISTS configs (
    env TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (env, key)
);

-- Создаем индекс для быстрого поиска по окружению
CREATE INDEX IF NOT EXISTS idx_configs_env ON configs(env);

-- Комментарии к таблице
COMMENT ON TABLE configs IS 'Таблица для хранения конфигураций различных окружений';
COMMENT ON COLUMN configs.env IS 'Имя окружения (production, development, staging и т.д.)';
COMMENT ON COLUMN configs.key IS 'Ключ конфигурации';
COMMENT ON COLUMN configs.value IS 'Значение конфигурации';
COMMENT ON COLUMN configs.updated_at IS 'Время последнего обновления';

