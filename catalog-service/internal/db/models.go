package db

import (
	"time"
	"github.com/shopspring/decimal"
)

func (UnitOfMeasure) TableName() string {
	return "units_of_measure"
}

type Product struct {
	ID 			string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name		string `gorm:"not null"`
	Active 		bool `gorm:"not null;default:true"`
	CreatedAt 	time.Time
}

type UnitOfMeasure struct {
	ID 					string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name 				string `gorm:"not null"`
	ConversionFactorKg 	decimal.Decimal `gorm:"type:numeric(12,4);not null"`
	CreatedAt 			time.Time
}

type Supplier struct {
	ID        string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string `gorm:"not null"`
	Document  string `gorm:"not null"`
	Address   string `gorm:"not null"`
	Active    bool   `gorm:"not null;default:true"`
	CreatedAt time.Time
}

