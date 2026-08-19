# Customer Functional Requirements

## Customer data

A customer has:

- `id`: system-generated UUID
- `first_name`: required
- `last_name`: required
- `email`: required and unique
- `status`: `active`, `blocked`, or `closed`
- `created_at`: system-generated creation time
- `updated_at`: system-generated last-update time

New customers have the `active` status.

## Operations

### Create a customer

`POST /customers`

- Requires `first_name`, `last_name`, and `email`.
- Returns the created customer.
- Returns `400 Bad Request` for invalid input.
- Returns `409 Conflict` when the email already belongs to another customer.

### List customers

`GET /customers`

- Returns all customers.

### Get a customer

`GET /customers/{customer_id}`

- Returns the requested customer.
- Returns `404 Not Found` when the customer does not exist.

### Update a customer

`PATCH /customers/{customer_id}`

- Can update `first_name`, `last_name`, `email`, or `status`.
- Requires at least one field to update.
- Returns the updated customer.
- Returns `400 Bad Request` for invalid input or status transitions.
- Returns `404 Not Found` when the customer does not exist.
- Returns `409 Conflict` when the email already belongs to another customer.

## Status transitions

Allowed transitions are:

- `active` to `blocked` or `closed`
- `blocked` to `active` or `closed`
- `closed` cannot transition to another status

## Out of scope

- Customer deletion
- Authentication and authorization
- KYC
- Addresses and phone numbers
