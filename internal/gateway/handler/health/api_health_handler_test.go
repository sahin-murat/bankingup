package health

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sahin-murat/bankingup/internal/database"
)

type fakeDatabase struct {
	ping func(context.Context) error
}

func (db fakeDatabase) Ping(ctx context.Context) error {
	return db.ping(ctx)
}

func TestHealthRoutes(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		database     database.Database
		wantStatus   int
		wantResponse string
	}{
		{
			name: "liveness is healthy",
			path: "/health",
			database: fakeDatabase{ping: func(context.Context) error {
				return errors.New("database should not be called")
			}},
			wantStatus:   fiber.StatusOK,
			wantResponse: "OK",
		},
		{
			name: "database is ready",
			path: "/health/ready",
			database: fakeDatabase{ping: func(context.Context) error {
				return nil
			}},
			wantStatus:   fiber.StatusOK,
			wantResponse: `{"database":"up","status":"ok"}`,
		},
		{
			name: "database is unavailable",
			path: "/health/ready",
			database: fakeDatabase{ping: func(context.Context) error {
				return errors.New("connection refused")
			}},
			wantStatus:   fiber.StatusServiceUnavailable,
			wantResponse: `{"database":"down","status":"unavailable"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			if err := RegisterHealthRoutes(app, tt.database); err != nil {
				t.Fatalf("RegisterHealthRoutes() error = %v", err)
			}

			response, err := app.Test(httptest.NewRequest(http.MethodGet, tt.path, http.NoBody))
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("io.ReadAll() error = %v", err)
			}

			if response.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.StatusCode, tt.wantStatus)
			}
			if string(body) != tt.wantResponse {
				t.Errorf("body = %q, want %q", string(body), tt.wantResponse)
			}
		})
	}
}

func TestHealthReadinessUsesTimeout(t *testing.T) {
	db := fakeDatabase{ping: func(ctx context.Context) error {
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("Ping() context has no deadline")
		}

		remaining := time.Until(deadline)
		if remaining <= 0 || remaining > databasePingTimeout {
			t.Fatalf("Ping() timeout = %v, want between 0 and %v", remaining, databasePingTimeout)
		}

		return nil
	}}

	app := fiber.New()
	if err := RegisterHealthRoutes(app, db); err != nil {
		t.Fatalf("RegisterHealthRoutes() error = %v", err)
	}

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/health/ready", http.NoBody))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Errorf("status = %d, want %d", response.StatusCode, fiber.StatusOK)
	}
}
