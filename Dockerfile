FROM golang:1.26.5-alpine3.24 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gateway ./cmd/gateway

FROM alpine:3.24

RUN addgroup -S app && adduser -S -G app app

COPY --from=build /out/gateway /usr/local/bin/gateway

USER app
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/gateway"]
