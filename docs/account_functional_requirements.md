# Account Functional Requirements

## Account data

An account has:

- `id`: system-generated UUID
- `customer_id`: owning customer
- `currency`: supported currency code
- `balance`: monetary balance
- `status`: `active`, `blocked`, `frozen`, or `closed`
- `created_at`: system-generated creation time
- `updated_at`: system-generated last-update time

New accounts have the `active` status. Currency cannot be changed after account
creation.

## Operations

### Create an account

`POST /accounts`

- Requires `customer_id` and `currency`.
- Accepts an optional `initial_balance` that defaults to zero.
- Allows account creation only for an `active` customer.
- Returns `400 Bad Request` for an unsupported currency, invalid precision, or a
  negative initial balance.
- Returns `404 Not Found` when the customer does not exist.
- Returns `409 Conflict` when the customer is not active.

### List accounts

`GET /accounts?customer_id={customer_id}`

- Returns accounts, optionally filtered by customer.

### Get an account

`GET /accounts/{account_id}`

- Returns the requested account.
- Returns `404 Not Found` when the account does not exist.

### Update an account

`PATCH /accounts/{account_id}`

- Can update only `status`.
- Returns `400 Bad Request` for an invalid status transition.
- Returns `404 Not Found` when the account does not exist.

## Status rules

- `active` allows normal operation.
- `blocked` prevents outgoing money movements.
- `frozen` prevents all money movements.
- `closed` is terminal and prevents all money movements.
- An account can be closed only when its balance is zero.
- `active`, `blocked`, and `frozen` can transition to each other or to `closed`.

## Balance rules

- Balance cannot be negative.
- Amounts use decimal strings in JSON and must follow the currency's allowed
  decimal places.
- Balance cannot be changed through the account update operation.
- Future money-movement operations will change balances atomically.

## Out of scope

- Account deletion
- Direct balance updates
- Deposits, withdrawals, and transfers
- Exchange rates and currency conversion
- Authentication and authorization
- Pagination
