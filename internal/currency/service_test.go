package currency

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	create    func(context.Context, Currency) (*Currency, error)
	list      func(context.Context) ([]Currency, error)
	getByCode func(context.Context, string) (*Currency, error)
}

func (r fakeRepository) Create(ctx context.Context, input Currency) (*Currency, error) {
	return r.create(ctx, input)
}

func (r fakeRepository) List(ctx context.Context) ([]Currency, error) {
	return r.list(ctx)
}

func (r fakeRepository) GetByCode(ctx context.Context, code string) (*Currency, error) {
	return r.getByCode(ctx, code)
}

func TestNewServiceRequiresRepository(t *testing.T) {
	service, err := NewService(nil)
	if err == nil {
		t.Fatal("NewService() error = nil, want an error")
	}
	if service != nil {
		t.Errorf("NewService() service = %v, want nil", service)
	}
}

func TestServiceCreate(t *testing.T) {
	want := Currency{Code: "CAD", Name: "Canadian Dollar", DecimalPlaces: 2}
	service, err := NewService(fakeRepository{
		create: func(_ context.Context, input Currency) (*Currency, error) {
			if input != want {
				t.Errorf("Create() input = %+v, want %+v", input, want)
			}
			return &input, nil
		},
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	created, err := service.Create(context.Background(), CreateInput{
		Code:          " cad ",
		Name:          " Canadian Dollar ",
		DecimalPlaces: 2,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if *created != want {
		t.Errorf("Create() = %+v, want %+v", *created, want)
	}
}

func TestServiceCreateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		input  CreateInput
		target error
	}{
		{
			name:   "invalid code",
			input:  CreateInput{Code: "CA1", Name: "Canadian Dollar", DecimalPlaces: 2},
			target: ErrInvalidCode,
		},
		{
			name:   "missing name",
			input:  CreateInput{Code: "CAD", DecimalPlaces: 2},
			target: ErrInvalidInput,
		},
		{
			name:   "negative decimal places",
			input:  CreateInput{Code: "CAD", Name: "Canadian Dollar", DecimalPlaces: -1},
			target: ErrInvalidInput,
		},
		{
			name:   "too many decimal places",
			input:  CreateInput{Code: "CAD", Name: "Canadian Dollar", DecimalPlaces: 5},
			target: ErrInvalidInput,
		},
	}

	service, err := NewService(fakeRepository{})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			created, err := service.Create(context.Background(), test.input)
			if !errors.Is(err, test.target) {
				t.Errorf("Create() error = %v, want %v", err, test.target)
			}
			if created != nil {
				t.Errorf("Create() = %+v, want nil", created)
			}
		})
	}
}

func TestServiceList(t *testing.T) {
	want := []Currency{{Code: "EUR", Name: "Euro", DecimalPlaces: 2}}
	service, err := NewService(fakeRepository{
		list: func(context.Context) ([]Currency, error) {
			return want, nil
		},
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	got, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("List() = %+v, want %+v", got, want)
	}
}

func TestServiceGetByCodeNormalizesCode(t *testing.T) {
	want := &Currency{Code: "EUR", Name: "Euro", DecimalPlaces: 2}
	service, err := NewService(fakeRepository{
		getByCode: func(_ context.Context, code string) (*Currency, error) {
			if code != "EUR" {
				t.Errorf("GetByCode() code = %q, want EUR", code)
			}
			return want, nil
		},
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	got, err := service.GetByCode(context.Background(), " eur ")
	if err != nil {
		t.Fatalf("GetByCode() error = %v", err)
	}
	if got != want {
		t.Errorf("GetByCode() = %+v, want %+v", got, want)
	}
}

func TestServiceGetByCodeRejectsInvalidCode(t *testing.T) {
	service, err := NewService(fakeRepository{})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	got, err := service.GetByCode(context.Background(), "EU1")
	if !errors.Is(err, ErrInvalidCode) {
		t.Errorf("GetByCode() error = %v, want %v", err, ErrInvalidCode)
	}
	if got != nil {
		t.Errorf("GetByCode() = %+v, want nil", got)
	}
}

func TestServiceGetByCodeRejectsUnsupportedCurrency(t *testing.T) {
	service, err := NewService(fakeRepository{
		getByCode: func(context.Context, string) (*Currency, error) {
			return nil, ErrNotFound
		},
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	got, err := service.GetByCode(context.Background(), "CAD")
	if !errors.Is(err, ErrUnsupportedCurrency) {
		t.Errorf("GetByCode() error = %v, want %v", err, ErrUnsupportedCurrency)
	}
	if got != nil {
		t.Errorf("GetByCode() = %+v, want nil", got)
	}
}
