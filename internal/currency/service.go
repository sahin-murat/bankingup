package currency

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidInput        = errors.New("invalid currency input")
	ErrInvalidCode         = errors.New("invalid currency code")
	ErrUnsupportedCurrency = errors.New("unsupported currency")
)

type CreateInput struct {
	Code          string
	Name          string
	DecimalPlaces int16
}

type Service interface {
	Create(context.Context, CreateInput) (*Currency, error)
	List(context.Context) ([]Currency, error)
	GetByCode(context.Context, string) (*Currency, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) (*service, error) {
	if repository == nil {
		return nil, errors.New("currency repository is required")
	}

	return &service{repository: repository}, nil
}

func (s *service) Create(ctx context.Context, input CreateInput) (*Currency, error) {
	code := strings.ToUpper(strings.TrimSpace(input.Code))
	if !validCode(code) {
		return nil, fmt.Errorf("%w: code must contain three letters", ErrInvalidCode)
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	if input.DecimalPlaces < 0 || input.DecimalPlaces > 4 {
		return nil, fmt.Errorf("%w: decimal_places must be between 0 and 4", ErrInvalidInput)
	}

	return s.repository.Create(ctx, Currency{
		Code:          code,
		Name:          name,
		DecimalPlaces: input.DecimalPlaces,
	})
}

func (s *service) List(ctx context.Context) ([]Currency, error) {
	return s.repository.List(ctx)
}

func (s *service) GetByCode(ctx context.Context, input string) (*Currency, error) {
	code := strings.ToUpper(strings.TrimSpace(input))
	if !validCode(code) {
		return nil, fmt.Errorf("%w: code must contain three letters", ErrInvalidCode)
	}

	found, err := s.repository.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedCurrency, code)
		}

		return nil, err
	}

	return found, nil
}

func validCode(code string) bool {
	if len(code) != 3 {
		return false
	}

	for _, character := range code {
		if character < 'A' || character > 'Z' {
			return false
		}
	}

	return true
}
