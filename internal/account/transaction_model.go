package account

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "deposit"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
	TransactionTypeTransfer   TransactionType = "transfer"
)

type Transaction struct {
	ID             uuid.UUID       `gorm:"column:id;type:uuid;default:uuidv7();primaryKey"`
	AccountID      uuid.UUID       `gorm:"column:account_id;type:uuid"`
	Type           TransactionType `gorm:"column:type;type:account_transaction_type"`
	Amount         decimal.Decimal `gorm:"column:amount;type:numeric(19,4)"`
	Currency       string          `gorm:"column:currency"`
	BalanceAfter   decimal.Decimal `gorm:"column:balance_after;type:numeric(19,4)"`
	IdempotencyKey uuid.UUID       `gorm:"column:idempotency_key;type:uuid"`
	TransferID     *uuid.UUID      `gorm:"column:transfer_id;type:uuid"`
	CreatedAt      time.Time       `gorm:"column:created_at;autoCreateTime:false;default:CURRENT_TIMESTAMP"`
}

func (Transaction) TableName() string {
	return "account_transactions"
}
