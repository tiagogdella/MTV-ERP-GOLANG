package db

import (
	"gorm.io/gorm"
)

type StockMovementRepository struct {
	db *gorm.DB
}

func NewStockMovementRepository(db *gorm.DB) *StockMovementRepository {
	return &StockMovementRepository{db:db}
}

func (s *StockMovementRepository) Create(movement *StockMovement) error {
	return s.db.Create(movement).Error
}

func (s *StockMovementRepository) ListByLot(lotID string) ([]StockMovement, error){
	var movements []StockMovement
	err := s.db.Where("lot_id = ?", lotID).Find(&movements).Error
	return movements, err
}