package db

import (
	"time"

	"github.com/shopspring/decimal"
)

type Purchase struct {
	ID            string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SupplierID    string          `gorm:"not null"`
	InvoiceNumber string          `gorm:"not null"`
	InvoiceDate   time.Time       `gorm:"type:date;not null"`
	InvoiceValue  decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	CreatedAt     time.Time
}

type PurchaseItem struct {
	ID         string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PurchaseID string          `gorm:"not null"`
	ProductID  string          `gorm:"not null"`
	UnitID     string          `gorm:"not null"`
	Quantity   decimal.Decimal `gorm:"type:numeric(12,4);not null"`
}
