CREATE TYPE NomineeRole AS ENUM ('member', 'seeder');
CREATE TYPE ProposalStatus AS ENUM ('created', 'approved', 'rejected');

CREATE TABLE IF NOT EXISTS proposals (
    id SERIAL PRIMARY KEY,
    nominator_id INTEGER NOT NULL,
    nominee_telegram_nickname VARCHAR NOT NULL,
    nominee_name VARCHAR NOT NULL,
    nominee_role NomineeRole NOT NULL,
    poll_id INTEGER NOT NULL,
    comment VARCHAR,
    created_at DATE NOT NULL DEFAULT CURRENT_DATE,
    finished_at DATE,
    status ProposalStatus NOT NULL DEFAULT 'created'
);