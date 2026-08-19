BEGIN;

CREATE TYPE account_transaction_type AS ENUM ('deposit', 'withdrawal', 'transfer');

CREATE TABLE account_transactions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    account_id UUID NOT NULL,
    type account_transaction_type NOT NULL,
    amount NUMERIC(19, 4) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    balance_after NUMERIC(19, 4) NOT NULL,
    idempotency_key UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT account_transactions_amount_positive
        CHECK (amount > 0),
    CONSTRAINT account_transactions_balance_not_negative
        CHECK (balance_after >= 0),
    CONSTRAINT account_transactions_account_fkey
        FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE RESTRICT,
    CONSTRAINT account_transactions_currency_fkey
        FOREIGN KEY (currency) REFERENCES currencies (code) ON DELETE RESTRICT,
    CONSTRAINT account_transactions_idempotency_key_unique
        UNIQUE (idempotency_key)
);

CREATE INDEX account_transactions_account_id_id_idx
    ON account_transactions (account_id, id);

COMMIT;
