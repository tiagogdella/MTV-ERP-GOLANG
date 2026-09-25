package db

import (
	"gorm.io/gorm"
)

type SupplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

func (r *SupplierRepository) Create(supplier *Supplier) error {
	return r.db.Create(supplier).Error
}

func (r *SupplierRepository) FindByID(id string) (*Supplier, error) {
	var supplier Supplier
	err := r.db.Where("id = ?", id).First(&supplier).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *SupplierRepository) ListActive() ([]Supplier, error) {
	var suppliers []Supplier
	err := r.db.Where("active = ?", true).Find(&suppliers).Error
	return suppliers, err
}

func (r *SupplierRepository) Deactivate(id string) error {
	return r.db.Model(&Supplier{}).Where("id = ?", id).Update("active", false).Error
}
