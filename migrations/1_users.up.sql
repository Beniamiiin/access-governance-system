CREATE TYPE UserRole AS ENUM ('guest', 'member', 'seeder');

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT,
    discord_id BIGINT,
    role UserRole NOT NULL DEFAULT 'guest',
    temp_proposal JSONB,
    telegram_state JSONB
);
