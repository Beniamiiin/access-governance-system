CREATE TYPE UserRole AS ENUM ('guest', 'member', 'seeder');

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    telegram_id BIGINT,
    telegram_nickname VARCHAR NOT NULL,
    discord_id BIGINT,
    role UserRole NOT NULL DEFAULT 'guest',
    backers_id VARCHAR[] NOT NULL DEFAULT '{}'::VARCHAR[],
    nominator_id INTEGER,
    temp_proposal JSONB,
    telegram_state JSONB
);
