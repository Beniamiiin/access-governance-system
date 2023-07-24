CREATE TYPE VoteType AS ENUM ('yes', 'no', 'acknowledge');

CREATE TABLE IF NOT EXISTS votes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    proposal_id INTEGER NOT NULL,
    type VoteType,
    comment VARCHAR,
    created_at DATE NOT NULL DEFAULT CURRENT_DATE
);
