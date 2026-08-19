BEGIN;

DROP INDEX account_transactions_transfer_id_idx;

ALTER TABLE account_transactions
    DROP CONSTRAINT account_transactions_transfer_fkey,
    DROP COLUMN transfer_id;

DROP TABLE transfers;

COMMIT;
