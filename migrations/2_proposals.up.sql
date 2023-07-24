CREATE TYPE NomineeRole AS ENUM ('member', 'seeder');
CREATE TYPE ProposalStatus AS ENUM ('created', 'approved', 'declined');

CREATE TABLE IF NOT EXISTS proposals (
    id SERIAL PRIMARY KEY,
    nominator_id INTEGER NOT NULL,
    nominee_telegram_nickname VARCHAR NOT NULL,
    nominee_telegram_id BIGINT NOT NULL,
    nominee_role NomineeRole NOT NULL,
    description VARCHAR NOT NULL,
    created_at DATE NOT NULL DEFAULT CURRENT_DATE,
    status ProposalStatus NOT NULL DEFAULT 'created'
);