# Transfer Functional Requirements

## Transfer data

A transfer has:

- `id`: system-generated UUID
- `source_account_id`: account sending money
- `destination_account_id`: account receiving money
- `amount`: positive monetary amount
- `currency`: currency shared by both accounts
- `source_balance_after`: source balance after the transfer
- `destination_balance_after`: destination balance after the transfer
- `idempotency_key`: client-provided UUID that prevents duplicate processing
- `created_at`: system-generated creation time

Transfer records are immutable.

## Operations

### Transfer money

`POST /transfers`

- Requires an `Idempotency-Key` header containing a UUID.
- Requires `source_account_id`, `destination_account_id`, and `amount`.
- Requires different source and destination accounts.
- Requires both accounts to use the same currency.
- Requires the source account to be `active`.
- Allows an `active` or `blocked` destination account.
- Rejects `frozen` or `closed` accounts.
- Requires sufficient funds in the source account.
- Returns the created transfer with `201 Created`.
- Returns the original transfer with `200 OK` for an idempotent replay.
- Returns `400 Bad Request` for invalid identifiers, amount, currency precision,
  or identical source and destination accounts.
- Returns `404 Not Found` when either account does not exist.
- Returns `409 Conflict` for incompatible currencies, restricted account
  statuses, insufficient funds, or conflicting idempotency-key reuse.

## Idempotency

- An idempotency key uniquely identifies one transfer request.
- Repeating the same transfer with the same key returns the original transfer
  and does not move money again.
- Reusing a key with different accounts or amount returns `409 Conflict`.
- Failed transfers do not reserve an idempotency key.

## Balance and consistency rules

- Transfer amounts are interpreted in the accounts' shared currency.
- Amounts must follow that currency's allowed decimal places.
- Both account rows are locked in deterministic ID order to prevent deadlocks.
- Both balance updates, the transfer record, and two account transaction records
  are created in one database transaction.
- One account transaction records the source debit and one records the
  destination credit; both use the `transfer` transaction type and reference
  the transfer.
- A transfer cannot produce a negative source balance.
- Any failure rolls back every balance and record change.

## Out of scope

- Transfers between different currencies
- Exchange rates and currency conversion
- Transfer cancellation or reversal
- Scheduled or recurring transfers
- Fees
- External payment processing
- Authentication and authorization
