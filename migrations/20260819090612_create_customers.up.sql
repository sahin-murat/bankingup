BEGIN;

CREATE TYPE customer_status AS ENUM ('active', 'blocked', 'closed');

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT NOT NULL,
    status customer_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT customers_first_name_not_blank
        CHECK (btrim(first_name) <> ''),
    CONSTRAINT customers_last_name_not_blank
        CHECK (btrim(last_name) <> ''),
    CONSTRAINT customers_email_not_blank
        CHECK (email = btrim(email) AND email <> '')
);

CREATE UNIQUE INDEX customers_email_unique
    ON customers (lower(email));

COMMIT;
