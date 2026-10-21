package db

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type LotRepository struct {
	db *gorm.DB
}

func NewLotsRepository(db *gorm.DB) *LotRepository {
	return &LotRepository{db:db}
}

func (l *LotRepository) Create(lot *Lot) error {
	return l.db.Create(lot).Error
}

func (l *LotRepository) FindByID(id string) (*Lot, error) {
	var lot Lot
	err := l.db.Where("id = ?", id).First(&lot).Error
	if err != nil {
		return nil, err
	}
	return &lot, nil
}

func (l *LotRepository) ListByProduct(productID string) ([]Lot, error) {
	var lots []Lot
	err := l.db.Where("product_id = ?", productID).Find(&lots).Error
	return lots, err
}

func (l *LotRepository) CreateWithEntry(lot *Lot, origin string) error {
	return l.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(lot).Error; err != nil {
			return err
		}

		entry := &StockMovement{
			LotID:      lot.ID,
			Type:       "entrada",
			QuantityKg: lot.QuantityKg,
			OccurredAt: lot.ReceivedAt,
			Origin:     origin,
		}
		return tx.Create(entry).Error
	})
}

type LotBalance struct {
	Lot       Lot             `gorm:"embedded"`
	BalanceKg decimal.Decimal
}

func (l *LotRepository) ListByProductWithBalance(productID string) ([]LotBalance, error) {
	var result []LotBalance
	err := l.db.Table("lots").
		Select("lots.*, COALESCE(SUM(stock_movements.quantity_kg), 0) AS balance_kg").
		Joins("LEFT JOIN stock_movements ON stock_movements.lot_id = lots.id").
		Where("lots.product_id = ?", productID).
		Group("lots.id").
		Order("lots.received_at").
		Scan(&result).Error
	return result, err
}