package account

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Status string

const (
	StatusActive  Status = "active"
	StatusBlocked Status = "blocked"
	StatusFrozen  Status = "frozen"
	StatusClosed  Status = "closed"
)

type Account struct {
	ID         uuid.UUID       `gorm:"column:id;type:uuid;default:uuidv7();primaryKey"`
	CustomerID uuid.UUID       `gorm:"column:customer_id;type:uuid"`
	Currency   string          `gorm:"column:currency"`
	Balance    decimal.Decimal `gorm:"column:balance;type:numeric(19,4)"`
	Status     Status          `gorm:"column:status;type:account_status;default:active"`
	CreatedAt  time.Time       `gorm:"column:created_at;autoCreateTime:false;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time       `gorm:"column:updated_at;autoUpdateTime:false;default:CURRENT_TIMESTAMP"`
}

func (Account) TableName() string {
	return "accounts"
}
