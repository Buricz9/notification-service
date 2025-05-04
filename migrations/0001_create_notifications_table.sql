-- Migrations
CREATE
EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE notifications
(
    id          UUID PRIMARY KEY     DEFAULT uuid_generate_v4(),
    user_id     UUID        NOT NULL,
    channel     TEXT        NOT NULL, -- "email" lub "push"
    payload     JSONB       NOT NULL, -- dowolne dane, które musimy wysłać
    status      TEXT        NOT NULL DEFAULT 'pending',
    retry_count INT         NOT NULL DEFAULT 0,
    error       TEXT,                 -- wiadomość o błędzie przy nieudanej próbie
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE notifications
    ADD COLUMN send_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  ADD COLUMN timezone TEXT NOT NULL DEFAULT 'UTC',
  ADD COLUMN priority INT NOT NULL DEFAULT 0;

