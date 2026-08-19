BEGIN;

CREATE TABLE transfers (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    source_account_id UUID NOT NULL,
    destination_account_id UUID NOT NULL,
    amount NUMERIC(19, 4) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    source_balance_after NUMERIC(19, 4) NOT NULL,
    destination_balance_after NUMERIC(19, 4) NOT NULL,
    idempotency_key UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT transfers_accounts_different
        CHECK (source_account_id <> destination_account_id),
    CONSTRAINT transfers_amount_positive
        CHECK (amount > 0),
    CONSTRAINT transfers_source_balance_not_negative
        CHECK (source_balance_after >= 0),
    CONSTRAINT transfers_destination_balance_not_negative
        CHECK (destination_balance_after >= 0),
    CONSTRAINT transfers_source_account_fkey
        FOREIGN KEY (source_account_id) REFERENCES accounts (id) ON DELETE RESTRICT,
    CONSTRAINT transfers_destination_account_fkey
        FOREIGN KEY (destination_account_id) REFERENCES accounts (id) ON DELETE RESTRICT,
    CONSTRAINT transfers_currency_fkey
        FOREIGN KEY (currency) REFERENCES currencies (code) ON DELETE RESTRICT,
    CONSTRAINT transfers_idempotency_key_unique
        UNIQUE (idempotency_key)
);

CREATE INDEX transfers_source_account_id_id_idx
    ON transfers (source_account_id, id);
CREATE INDEX transfers_destination_account_id_id_idx
    ON transfers (destination_account_id, id);

ALTER TABLE account_transactions
    ADD COLUMN transfer_id UUID,
    ADD CONSTRAINT account_transactions_transfer_fkey
        FOREIGN KEY (transfer_id) REFERENCES transfers (id) ON DELETE RESTRICT;

CREATE INDEX account_transactions_transfer_id_idx
    ON account_transactions (transfer_id)
    WHERE transfer_id IS NOT NULL;

COMMIT;
