# Bankingup contributor guide

## Project overview

`bankingup` is a Go 1.26 API built with Fiber v3. The executable is
`cmd/gateway`; application code belongs under `internal/`.

## Layout

- `cmd/gateway/main.go` wires configuration and starts the HTTP server.
- `internal/config` owns environment-based configuration. Extend its `Config`
  interface when a new setting is needed.
- `internal/gateway` owns Fiber initialization and route registration.
- `internal/gateway/handler/<feature>` contains each feature's route setup and
  handlers.
- `internal/database` provides the GORM/PostgreSQL connection wrapper.
- `migrations` contains timestamped `golang-migrate` SQL migration pairs.
- `test/postman` contains the API smoke-test collection.

## Local development

Format and verify Go changes before handing them off:

```sh
gofmt -w $(rg --files -g '*.go' cmd internal)
go test ./...
go build -o /tmp/bankingup-gateway ./cmd/gateway
```

Run the service directly with:

```sh
go run ./cmd/gateway
```

The default address is `:8080`, and the default host-machine database URL is:

```text
postgresql://postgres:postgres@localhost:5432/postgres
```

To run the full stack, use the Compose file (Docker Compose or Podman Compose):

```sh
podman compose up --build
```

Within Compose, the gateway reaches PostgreSQL at
`postgres://bankingup:bankingup@postgres:5432/bankingup?sslmode=disable`.
From the host, use
`postgres://bankingup:bankingup@localhost:5432/bankingup?sslmode=disable`
when the Compose database is running.

## Implementation conventions

- Keep packages small and feature-oriented; register every new endpoint from
  `gateway.DefineRoutes`.
- Follow RESTful API naming conventions: use plural resource nouns, lowercase
  paths, standard HTTP methods, and avoid verbs in endpoint URLs unless
  modeling an action is unavoidable.
- Pass dependencies through narrow interfaces (as `config.Config` and
  `database.Config` do) so handlers remain testable.
- Wrap returned errors with actionable context using `%w`; do not log an error
  and return it again at lower layers.
- Do not hard-code credentials outside configuration defaults or Compose
  development settings. Read runtime values from environment variables.
- Add table-driven unit tests for new behavior. Keep external service tests
  separate from unit tests.
- Run `gofmt` on every changed Go file; preserve the existing package naming
  and import style.

## API checks

The Postman collection uses `http://localhost:8080` by default. Update it when
you add or change public endpoints, including response assertions where useful.
Group Postman requests into folders by entity or domain, such as `Customers`;
keep cross-cutting technical endpoints in a dedicated folder such as `Health`.

## Database migrations

- Manage the schema with `golang-migrate`; do not use GORM `AutoMigrate`.
- Store timestamped `.up.sql` and `.down.sql` files in `migrations/` and create
  them with `migrate create -ext sql -dir migrations -format 20060102150405
  <name>`.
- Never edit a migration after it has been merged or applied. Add a new
  migration instead.
- Wrap normal migrations in `BEGIN` and `COMMIT`. Put operations that cannot run
  in a transaction, such as `CREATE INDEX CONCURRENTLY`, in a separate
  migration.
- Keep development seed data out of schema migrations. Define constraints,
  foreign keys, and indexes explicitly.
- Never automate `migrate force`; inspect and repair dirty migration state
  manually.
- Prefer forward fixes in production and use expand-and-contract migrations for
  breaking schema changes.
- Run migrations once before the gateway starts. Gateway replicas must not run
  migrations during application startup.

## Application architecture

- Domain request flow should normally be `handler -> service -> repository ->
  database`.
- Handlers own HTTP concerns only: request parsing, basic input validation,
  invoking a use case, mapping errors, and writing responses.
- Services own business rules, authorization decisions, coordination, and
  transaction boundaries.
- Repositories own persistence and hide GORM and PostgreSQL details.
- Technical endpoints such as health checks may use infrastructure directly.
- Do not create an empty pass-through service when it has no business
  responsibility.
- **Feature composition:** Each domain feature's `routers.go` is its composition
  point. Route registration receives the Fiber application and required shared
  infrastructure, constructs one repository and one service for the feature,
  constructs the operation-specific handlers, and registers their routes.
  Route registration runs once during application startup.
- **Handler dependencies:** Handler constructors receive only the narrow service
  interface required by that operation. Handlers must not construct repositories
  or services, access GORM directly, or receive dependencies unrelated to their
  operation. Reuse the same feature service instance across its handlers.
- **Technical handlers:** Technical endpoints such as health checks may receive
  a narrow infrastructure interface directly when that infrastructure operation
  is the complete use case.
- **Handler organization:** For domain resources, implement each operation as a
  separate handler type and file, such as `create_customer_handler.go`,
  `update_customer_handler.go`, and `list_customers_handler.go`. Give each
  handler only the narrow service interface required by that operation and keep
  its tests in a corresponding test file. Keep route registration in one
  feature-level `routers.go` file. Do not combine all CRUD operations into one
  large handler, and do not create a separate package for every operation.

Feature construction follows this flow:

```text
gateway
  -> RegisterCustomerRoutes(app, gormDB)
      -> customer.NewRepository(gormDB)
      -> customer.NewService(repository)
      -> NewCreateCustomerHandler(service)
      -> NewUpdateCustomerHandler(service)
      -> register routes
```

## Practical Rules for Agents

- **Staff-level engineering judgment:** Work as a staff-level software engineer. Make technically sound decisions that consider correctness, maintainability, consistency, and meaningful risks. Prefer the simplest solution that fully meets the requirements, and do not introduce production-scale complexity, abstractions, or infrastructure without a demonstrated need.

- **Proposal before implementation:** Before implementing any change, present the proposed approach, scope, and important decisions to the user and wait for explicit approval. Read-only inspection and planning are allowed before approval; modifying files, adding dependencies, changing databases, or starting stateful services is not.

- **No over-engineering:** Keep solutions minimal and straightforward. Do not introduce abstractions, patterns, or layers beyond what the task requires. Match the complexity of the solution to the complexity of the problem.

- **Avoid simple mistakes and pitfalls:** Double-check for typos, wrong import paths, missing error handling, incorrect config keys, and other trivial errors before proposing changes. Review the diff once more before presenting it.

- **No unrequested behavior:** Do not add features, refactors, optimizations, or "nice-to-haves" the user did not ask for. Stick strictly to the scope of the request.

- **No unauthorized refactoring:** Never refactor code unless the user explicitly approves the refactoring. Do not modify files, functions, behavior, formatting, or unrelated code outside the requested scope. If an out-of-scope issue is discovered, report it separately without changing it.

- **Keep documentation concise:** Write human-facing documentation, especially `README.md`, in clear and direct language. Include only information needed to understand, run, or use the project. Avoid excessive explanation, repetition, and unnecessary sections.

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [RFC 9110: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html)
- [Microsoft Azure REST API Guidelines](https://github.com/microsoft/api-guidelines/blob/vNext/azure/Guidelines.md)
- [GORM documentation](https://gorm.io/docs/)
- [PostgreSQL Docker Official Image](https://hub.docker.com/_/postgres)
- [golang-migrate](https://github.com/golang-migrate/migrate)
