BEGIN;

CREATE TYPE account_status AS ENUM ('active', 'blocked', 'frozen', 'closed');

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    customer_id UUID NOT NULL,
    currency VARCHAR(3) NOT NULL,
    balance NUMERIC(19, 4) NOT NULL DEFAULT 0,
    status account_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT accounts_balance_not_negative
        CHECK (balance >= 0),
    CONSTRAINT accounts_customer_fkey
        FOREIGN KEY (customer_id) REFERENCES customers (id) ON DELETE RESTRICT,
    CONSTRAINT accounts_currency_fkey
        FOREIGN KEY (currency) REFERENCES currencies (code) ON DELETE RESTRICT
);

CREATE INDEX accounts_customer_id_idx ON accounts (customer_id);

COMMIT;
