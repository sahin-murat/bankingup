# bankingup
A Golang banking API built wih Fiber.

## Docker

Build and run the service:

```sh
docker build -t banking-up .
docker run --rm -p 8080:8080 banking-up
```

## Podman

As an alternative to Docker, build and run the same container with Podman:

```sh
podman build -t banking-up .
podman run --rm -p 8080:8080 banking-up
```
