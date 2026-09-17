package db

import (
	"time"

	"github.com/shopspring/decimal"
)

type Lot struct {
	ID             string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProductID      string          `gorm:"not null"`
	PurchaseItemID string          `gorm:"not null"`
	Safra          string          `gorm:"not null"`
	QuantityKg     decimal.Decimal `gorm:"type:numeric(12,4);not null"`
	ReceivedAt     time.Time       `gorm:"type:date;not null"`
}

type StockMovement struct {
	ID         string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	LotID      string          `gorm:"not null"`
	Type       string          `gorm:"not null"`
	QuantityKg decimal.Decimal `gorm:"type:numeric(12,4);not null"`
	OccurredAt time.Time       `gorm:"not null"`
	Origin     string          `gorm:"not null"`
}