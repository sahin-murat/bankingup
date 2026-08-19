package customer

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusActive  Status = "active"
	StatusBlocked Status = "blocked"
	StatusClosed  Status = "closed"
)

type Customer struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;default:uuidv7();primaryKey"`
	FirstName string    `gorm:"column:first_name"`
	LastName  string    `gorm:"column:last_name"`
	Email     string    `gorm:"column:email"`
	Status    Status    `gorm:"column:status;type:customer_status;default:active"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime:false;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime:false;default:CURRENT_TIMESTAMP"`
}

func (Customer) TableName() string {
	return "customers"
}
