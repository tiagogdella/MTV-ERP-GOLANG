package db

import (
	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(product *Product) error {
	return r.db.Create(product).Error
}

func (r *ProductRepository) ListActive() ([]Product, error) {
	var products []Product
	err := r.db.Where("active = ?", true).Find(&products).Error
	return products, err
}

func (r *ProductRepository) Deactivate(id string) error {
	return r.db.Model(&Product{}).Where("id = ?", id).Update("active", false).Error
}