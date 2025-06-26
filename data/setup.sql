CREATE TYPE permission AS ENUM ('dumps_read', 'dumps_write');

CREATE TABLE IF NOT EXISTS users(
    id                  UUID PRIMARY KEY,
    email               VARCHAR(255) UNIQUE NOT NULL,
    password            VARCHAR(255),
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP,
    login_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    firstname           VARCHAR(255),
    surname             VARCHAR(255),
    avatar_url          VARCHAR(255),
    oauth_provider      VARCHAR(255),
    oauth_provider_id   VARCHAR(255),
    oauth_access_token  VARCHAR(255),
    oauth_refresh_token VARCHAR(255),
    oauth_expires_at    TIMESTAMP
);
CREATE TABLE IF NOT EXISTS permissions(
    id          UUID PRIMARY KEY,
    user_id     UUID REFERENCES users(id),
    permission  permission NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS dumps(
    id          UUID PRIMARY KEY,
    description TEXT NOT NULL,
    user_id     UUID REFERENCES users(id),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP
);
CREATE TABLE IF NOT EXISTS dump_entries(
    id           UUID PRIMARY KEY,
    dumps_id     UUID REFERENCES dumps(id),
    amount       SMALLINT,
    occurred_at  TIMESTAMPTZ DEFAULT NOW()
);