# Account Transaction Functional Requirements

## Transaction data

An account transaction has:

- `id`: system-generated UUID
- `account_id`: affected account
- `type`: `deposit`, `withdrawal`, or `transfer`
- `amount`: positive monetary amount
- `currency`: account currency at the time of the transaction
- `balance_after`: account balance after the transaction
- `idempotency_key`: UUID that prevents duplicate processing; client-provided
  for API movements and system-generated for opening deposits and transfer
  ledger entries
- `created_at`: system-generated creation time

Transaction type is stored as a PostgreSQL enum. Transfer behavior is defined
in [Transfer Functional Requirements](transfer_functional_requirements.md).

## Operations

### Deposit money

`POST /accounts/{account_id}/deposits`

- Requires an `Idempotency-Key` header containing a UUID.
- Requires `amount` as a positive decimal string.
- Allows deposits into `active` and `blocked` accounts.
- Returns the created transaction with `201 Created`.
- Returns `400 Bad Request` for a missing or invalid idempotency key, invalid
  amount, or invalid currency precision.
- Returns `404 Not Found` when the account does not exist.
- Returns `409 Conflict` when the account is `frozen` or `closed`.

### Withdraw money

`POST /accounts/{account_id}/withdrawals`

- Requires an `Idempotency-Key` header containing a UUID.
- Requires `amount` as a positive decimal string.
- Allows withdrawals only from `active` accounts.
- Returns the created transaction with `201 Created`.
- Returns `400 Bad Request` for a missing or invalid idempotency key, invalid
  amount, or invalid currency precision.
- Returns `404 Not Found` when the account does not exist.
- Returns `409 Conflict` when the account is `blocked`, `frozen`, or `closed`,
  or when the account has insufficient funds.

## Idempotency

- An idempotency key uniquely identifies one money-movement request.
- Repeating the same request with the same key returns the original transaction
  with `200 OK` and does not change the balance again.
- Reusing a key with a different account, operation, or amount returns
  `409 Conflict`.
- Failed requests do not reserve an idempotency key.

## Balance and consistency rules

- Transaction amounts must follow the account currency's allowed decimal
  places.
- Deposit and withdrawal amounts are always interpreted in the account's
  currency.
- Balance updates and transaction creation occur in one database transaction.
- The account row is locked while its balance is checked and updated.
- A withdrawal cannot produce a negative balance.
- Any failure rolls back both the balance update and transaction creation.
- Transaction records are immutable.
- A non-zero initial account balance creates a `deposit` transaction in the same
  database transaction as account creation.

## Out of scope

- Transaction history endpoints
- Transaction reversal or deletion
- Fees and interest
- Caller-supplied currency selection or validation; movement requests do not
  accept a currency field
- Authentication and authorization
