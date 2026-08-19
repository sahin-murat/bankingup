package customer

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	create  func(context.Context, Customer) (*Customer, error)
	getByID func(context.Context, uuid.UUID) (*Customer, error)
	list    func(context.Context) ([]Customer, error)
	update  func(context.Context, Customer) (*Customer, error)
}

func (r fakeRepository) Create(ctx context.Context, input Customer) (*Customer, error) {
	return r.create(ctx, input)
}

func (r fakeRepository) GetByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
	return r.getByID(ctx, id)
}

func (r fakeRepository) List(ctx context.Context) ([]Customer, error) {
	return r.list(ctx)
}

func (r fakeRepository) Update(ctx context.Context, input Customer) (*Customer, error) {
	return r.update(ctx, input)
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
	want := Customer{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     "ada@example.com",
		Status:    StatusActive,
	}

	repository := fakeRepository{
		create: func(_ context.Context, input Customer) (*Customer, error) {
			if input != want {
				t.Errorf("Create() input = %+v, want %+v", input, want)
			}
			return &input, nil
		},
	}
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	created, err := service.Create(context.Background(), CreateInput{
		FirstName: "  Ada ",
		LastName:  " Lovelace  ",
		Email:     " ADA@EXAMPLE.COM ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if *created != want {
		t.Errorf("Create() = %+v, want %+v", *created, want)
	}
}

func TestServiceCreateRejectsMissingFields(t *testing.T) {
	tests := []struct {
		name  string
		input CreateInput
	}{
		{
			name:  "first name",
			input: CreateInput{LastName: "Lovelace", Email: "ada@example.com"},
		},
		{
			name:  "last name",
			input: CreateInput{FirstName: "Ada", Email: "ada@example.com"},
		},
		{
			name:  "email",
			input: CreateInput{FirstName: "Ada", LastName: "Lovelace"},
		},
	}

	service, err := NewService(fakeRepository{})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			created, err := service.Create(context.Background(), test.input)
			if !errors.Is(err, ErrInvalidInput) {
				t.Errorf("Create() error = %v, want %v", err, ErrInvalidInput)
			}
			if created != nil {
				t.Errorf("Create() customer = %+v, want nil", created)
			}
		})
	}
}

func TestServiceUpdate(t *testing.T) {
	id := uuid.New()
	existing := Customer{
		ID:        id,
		FirstName: "Ada",
		LastName:  "Byron",
		Email:     "ada@example.com",
		Status:    StatusActive,
	}
	firstName := "  Augusta Ada "
	email := " ADA.LOVELACE@EXAMPLE.COM "
	status := StatusBlocked
	want := existing
	want.FirstName = "Augusta Ada"
	want.Email = "ada.lovelace@example.com"
	want.Status = StatusBlocked

	repository := fakeRepository{
		getByID: func(_ context.Context, requestedID uuid.UUID) (*Customer, error) {
			if requestedID != id {
				t.Errorf("GetByID() id = %v, want %v", requestedID, id)
			}
			found := existing
			return &found, nil
		},
		update: func(_ context.Context, input Customer) (*Customer, error) {
			if input != want {
				t.Errorf("Update() input = %+v, want %+v", input, want)
			}
			return &input, nil
		},
	}
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	updated, err := service.Update(context.Background(), UpdateInput{
		ID:        id,
		FirstName: &firstName,
		Email:     &email,
		Status:    &status,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if *updated != want {
		t.Errorf("Update() = %+v, want %+v", *updated, want)
	}
}

func TestServiceUpdateRequiresChanges(t *testing.T) {
	service, err := NewService(fakeRepository{})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	updated, err := service.Update(context.Background(), UpdateInput{ID: uuid.New()})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("Update() error = %v, want %v", err, ErrInvalidInput)
	}
	if updated != nil {
		t.Errorf("Update() customer = %+v, want nil", updated)
	}
}

func TestStatusTransitions(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		want bool
	}{
		{name: "active to blocked", from: StatusActive, to: StatusBlocked, want: true},
		{name: "active to closed", from: StatusActive, to: StatusClosed, want: true},
		{name: "blocked to active", from: StatusBlocked, to: StatusActive, want: true},
		{name: "blocked to closed", from: StatusBlocked, to: StatusClosed, want: true},
		{name: "same status", from: StatusActive, to: StatusActive, want: true},
		{name: "closed to active", from: StatusClosed, to: StatusActive, want: false},
		{name: "unknown target", from: StatusActive, to: Status("unknown"), want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := canTransition(test.from, test.to); got != test.want {
				t.Errorf("canTransition(%q, %q) = %v, want %v", test.from, test.to, got, test.want)
			}
		})
	}
}

func TestServicePassThroughQueries(t *testing.T) {
	id := uuid.New()
	wantCustomer := &Customer{ID: id}
	wantCustomers := []Customer{{ID: id}}
	repository := fakeRepository{
		getByID: func(context.Context, uuid.UUID) (*Customer, error) {
			return wantCustomer, nil
		},
		list: func(context.Context) ([]Customer, error) {
			return wantCustomers, nil
		},
	}
	service, err := NewService(repository)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	gotCustomer, err := service.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if gotCustomer != wantCustomer {
		t.Errorf("GetByID() = %+v, want %+v", gotCustomer, wantCustomer)
	}

	gotCustomers, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if !reflect.DeepEqual(gotCustomers, wantCustomers) {
		t.Errorf("List() = %+v, want %+v", gotCustomers, wantCustomers)
	}
}
