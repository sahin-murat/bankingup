package transfer

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Transfer struct {
	ID                      uuid.UUID       `gorm:"column:id;type:uuid;default:uuidv7();primaryKey"`
	SourceAccountID         uuid.UUID       `gorm:"column:source_account_id;type:uuid"`
	DestinationAccountID    uuid.UUID       `gorm:"column:destination_account_id;type:uuid"`
	Amount                  decimal.Decimal `gorm:"column:amount;type:numeric(19,4)"`
	Currency                string          `gorm:"column:currency"`
	SourceBalanceAfter      decimal.Decimal `gorm:"column:source_balance_after;type:numeric(19,4)"`
	DestinationBalanceAfter decimal.Decimal `gorm:"column:destination_balance_after;type:numeric(19,4)"`
	IdempotencyKey          uuid.UUID       `gorm:"column:idempotency_key;type:uuid"`
	CreatedAt               time.Time       `gorm:"column:created_at;autoCreateTime:false;default:CURRENT_TIMESTAMP"`
}

func (Transfer) TableName() string {
	return "transfers"
}
