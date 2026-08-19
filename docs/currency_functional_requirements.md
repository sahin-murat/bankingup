# Currency Functional Requirements

## Currency data

A currency has:

- `code`: three-letter uppercase currency code
- `name`: display name
- `decimal_places`: number of decimal places allowed for amounts

The initially supported currencies are `EUR`, `USD`, `GBP`, `TRY`, and `JPY`.

## Operations

### Create a currency

`POST /currencies`

- Requires `code`, `name`, and `decimal_places`.
- Normalizes `code` to uppercase and trims text fields.
- Requires a three-letter alphabetic code and a non-empty name.
- Requires `decimal_places` to be between `0` and `4`.
- Returns the created currency.
- Returns `400 Bad Request` for invalid input.
- Returns `409 Conflict` when the currency code already exists.

### List currencies

`GET /currencies`

- Returns all supported currencies ordered by code.

## Rules

- The initial supported currencies are created through a database migration.
- Additional supported currencies can be created through the API.
- Currency codes cannot be changed.

## Out of scope

- Updating or deleting currencies through the API
- Exchange rates
- Currency conversion
