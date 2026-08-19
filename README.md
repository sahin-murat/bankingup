# bankingup
A Golang banking API built wih Fiber.

## Run with Compose

Start the gateway and PostgreSQL with Docker or Podman:

```sh
docker compose up --build
# or
podman compose up --build
```

Stop the services with:

```sh
docker compose down
# or
podman compose down
```

The gateway runs at `http://localhost:8080`. Use `GET /health` for liveness
and `GET /health/ready` for database readiness. PostgreSQL is exposed on port
`5432` and stores its data in the `postgres-data` volume.

Override the default ports with `APP_PORT` and `POSTGRES_PORT`.

## Postman

The Postman collection is located at [`test/postman/bankingup.postman_collection.json`](test/postman/bankingup.postman_collection.json). It uses `http://localhost:8080` as the default `baseUrl` and includes response assertions.
