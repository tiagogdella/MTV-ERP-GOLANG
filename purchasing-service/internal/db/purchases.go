package db

import (
	"gorm.io/gorm"
)

type PurchaseRepository struct {
	db *gorm.DB
}

func NewPurchaseRepository(db *gorm.DB) *PurchaseRepository {
	return &PurchaseRepository{db: db}
}

func (r *PurchaseRepository) Create(purchase *Purchase, items []PurchaseItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(purchase).Error; err != nil {
			return err
		}

		for i := range items {
			items[i].PurchaseID = purchase.ID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
