package db

import (
	"gorm.io/gorm"
)

type UnitOfMeasureRepository struct {
	db *gorm.DB
}

func NewUnitOfMeasureRepository(db *gorm.DB) *UnitOfMeasureRepository {
	return &UnitOfMeasureRepository{db: db}
}

func (r *UnitOfMeasureRepository) Create(unitOfMeasure *UnitOfMeasure) error {
	return r.db.Create(unitOfMeasure).Error
}

func (r *UnitOfMeasureRepository) FindByID(id string) (*UnitOfMeasure, error) {
	var unitOfMeasure UnitOfMeasure
	err := r.db.Where("id = ?", id).First(&unitOfMeasure).Error
	if err != nil {
		return nil, err
	}
	return &unitOfMeasure, nil
}