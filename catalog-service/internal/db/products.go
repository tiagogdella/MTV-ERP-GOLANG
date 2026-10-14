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

func (r *ProductRepository) FindByID(id string) (*Product, error) {
	var product Product
	err := r.db.Where("id = ?", id).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) ListActive() ([]Product, error) {
	var products []Product
	err := r.db.Where("active = ?", true).Find(&products).Error
	return products, err
}

// Item 2.3 === Erro de não encontrado != falha de banco ===
func (r *ProductRepository) Deactivate(id string) error {
	result := r.db.Model(&Product{}).Where("id = ?", id).Update("active", false)

	if result.Error != nil {
		return result.Error // Falha banco
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // Não encontrado
	}
	return nil
}
