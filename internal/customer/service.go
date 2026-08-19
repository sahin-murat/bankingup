package customer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidInput            = errors.New("invalid customer input")
	ErrInvalidStatusTransition = errors.New("invalid customer status transition")
)

type CreateInput struct {
	FirstName string
	LastName  string
	Email     string
}

type UpdateInput struct {
	ID        uuid.UUID
	FirstName *string
	LastName  *string
	Email     *string
	Status    *Status
}

type Service interface {
	Create(context.Context, CreateInput) (*Customer, error)
	GetByID(context.Context, uuid.UUID) (*Customer, error)
	List(context.Context) ([]Customer, error)
	Update(context.Context, UpdateInput) (*Customer, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) (*service, error) {
	if repository == nil {
		return nil, errors.New("customer repository is required")
	}

	return &service{repository: repository}, nil
}

func (s *service) Create(ctx context.Context, input CreateInput) (*Customer, error) {
	firstName := strings.TrimSpace(input.FirstName)
	if firstName == "" {
		return nil, fmt.Errorf("%w: first_name is required", ErrInvalidInput)
	}

	lastName := strings.TrimSpace(input.LastName)
	if lastName == "" {
		return nil, fmt.Errorf("%w: last_name is required", ErrInvalidInput)
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" {
		return nil, fmt.Errorf("%w: email is required", ErrInvalidInput)
	}

	return s.repository.Create(ctx, Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Status:    StatusActive,
	})
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context) ([]Customer, error) {
	return s.repository.List(ctx)
}

func (s *service) Update(ctx context.Context, input UpdateInput) (*Customer, error) {
	if input.FirstName == nil &&
		input.LastName == nil &&
		input.Email == nil &&
		input.Status == nil {
		return nil, fmt.Errorf("%w: at least one field is required", ErrInvalidInput)
	}

	existing, err := s.repository.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.FirstName != nil {
		firstName := strings.TrimSpace(*input.FirstName)
		if firstName == "" {
			return nil, fmt.Errorf("%w: first_name is required", ErrInvalidInput)
		}
		existing.FirstName = firstName
	}

	if input.LastName != nil {
		lastName := strings.TrimSpace(*input.LastName)
		if lastName == "" {
			return nil, fmt.Errorf("%w: last_name is required", ErrInvalidInput)
		}
		existing.LastName = lastName
	}

	if input.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*input.Email))
		if email == "" {
			return nil, fmt.Errorf("%w: email is required", ErrInvalidInput)
		}
		existing.Email = email
	}

	if input.Status != nil {
		if !canTransition(existing.Status, *input.Status) {
			return nil, fmt.Errorf(
				"%w: %s to %s",
				ErrInvalidStatusTransition,
				existing.Status,
				*input.Status,
			)
		}
		existing.Status = *input.Status
	}

	return s.repository.Update(ctx, *existing)
}

func canTransition(from Status, to Status) bool {
	if from == to {
		return true
	}

	switch from {
	case StatusActive:
		return to == StatusBlocked || to == StatusClosed
	case StatusBlocked:
		return to == StatusActive || to == StatusClosed
	case StatusClosed:
		return false
	default:
		return false
	}
}

var _ Service = (*service)(nil)
