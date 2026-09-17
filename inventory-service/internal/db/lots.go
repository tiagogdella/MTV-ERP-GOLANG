package db

import (
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
