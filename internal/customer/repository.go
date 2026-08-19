package customer

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const customersEmailUniqueConstraint = "customers_email_unique"

var (
	ErrNotFound           = errors.New("customer not found")
	ErrEmailAlreadyExists = errors.New("customer email already exists")
)

type Repository interface {
	Create(context.Context, Customer) (*Customer, error)
	GetByID(context.Context, uuid.UUID) (*Customer, error)
	List(context.Context) ([]Customer, error)
	Update(context.Context, Customer) (*Customer, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) (*repository, error) {
	if db == nil {
		return nil, errors.New("gorm database is required")
	}

	return &repository{db: db}, nil
}

func (r *repository) Create(ctx context.Context, input Customer) (*Customer, error) {
	created := input

	if err := r.db.WithContext(ctx).Create(&created).Error; err != nil {
		return nil, repositoryError("create customer", err)
	}

	return &created, nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
	var found Customer

	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&found).Error; err != nil {
		return nil, repositoryError("get customer by id", err)
	}

	return &found, nil
}

func (r *repository) List(ctx context.Context) ([]Customer, error) {
	customers := make([]Customer, 0)

	if err := r.db.WithContext(ctx).Find(&customers).Error; err != nil {
		return nil, repositoryError("list customers", err)
	}

	return customers, nil
}

func (r *repository) Update(ctx context.Context, input Customer) (*Customer, error) {
	result := r.db.WithContext(ctx).
		Model(&Customer{}).
		Where("id = ?", input.ID).
		Updates(map[string]any{
			"first_name": input.FirstName,
			"last_name":  input.LastName,
			"email":      input.Email,
			"status":     input.Status,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})

	if result.Error != nil {
		return nil, repositoryError("update customer", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("update customer: %w", ErrNotFound)
	}

	return r.GetByID(ctx, input.ID)
}

func repositoryError(operation string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%s: %w", operation, ErrNotFound)
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) &&
		postgresError.Code == "23505" &&
		postgresError.ConstraintName == customersEmailUniqueConstraint {
		return fmt.Errorf("%s: %w", operation, ErrEmailAlreadyExists)
	}

	return fmt.Errorf("%s: %w", operation, err)
}
