package grpcserver

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"mtv-erp/inventory-service/internal/db"
	inventoryv1 "mtv-erp/inventory-service/internal/pb/inventory/v1"
)

type Server struct {
	inventoryv1.UnimplementedInventoryServiceServer
	lotRepo      *db.LotRepository
	movementRepo *db.StockMovementRepository
}

func NewServer(lotRepo *db.LotRepository, movementRepo *db.StockMovementRepository) *Server {
	return &Server{lotRepo: lotRepo, movementRepo: movementRepo}
}

func (s *Server) CreateLot(ctx context.Context, req *inventoryv1.CreateLotRequest) (*inventoryv1.CreateLotResponse, error) {
	if err := validateID(req.ProductId, "product_id"); err != nil {
		return nil, err
	}
	if err := validateID(req.PurchaseItemId, "purchase_item_id"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Safra) == "" {
		return nil, status.Error(codes.InvalidArgument, "safra é obrigatória")
	}

	quantity, err := decimal.NewFromString(req.QuantityKg)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "quantity_kg %q não é um número válido", req.QuantityKg)
	}
	if !quantity.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, "quantity_kg deve ser maior que zero")
	}

	receivedAt, err := time.Parse("2006-01-02", req.ReceivedAt)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "received_at %q inválido, use AAAA-MM-DD", req.ReceivedAt)
	}

	lot := &db.Lot{
		ProductID:      req.ProductId,
		PurchaseItemID: req.PurchaseItemId,
		Safra:          req.Safra,
		QuantityKg:     quantity,
		ReceivedAt:     receivedAt,
	}

	if err := s.lotRepo.CreateWithEntry(lot, "compra"); err != nil {
		return nil, status.Errorf(codes.Internal, "erro ao criar lote: %v", err)
	}

	return &inventoryv1.CreateLotResponse{Lot: toPBLot(lot)}, nil
}

func (s *Server) RegisterMovement(ctx context.Context, req *inventoryv1.RegisterMovementRequest) (*inventoryv1.RegisterMovementResponse, error) {
	if err := validateID(req.LotId, "lot_id"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Type) == "" || strings.TrimSpace(req.Origin) == "" {
		return nil, status.Error(codes.InvalidArgument, "type e origin são obrigatórios")
	}

	quantity, err := decimal.NewFromString(req.QuantityKg)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "quantity_kg %q não é um número válido", req.QuantityKg)
	}
	if quantity.IsZero() {
		return nil, status.Error(codes.InvalidArgument, "quantity_kg não pode ser zero")
	}

	occurredAt, err := time.Parse(time.RFC3339, req.OccurredAt)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "occurred_at %q inválido, use RFC3339", req.OccurredAt)
	}

	// regra de domínio explícita: sem lote não existe estoque (não é só a FK do banco)
	if _, err := s.findLot(req.LotId); err != nil {
		return nil, err
	}

	movement := &db.StockMovement{
		LotID:      req.LotId,
		Type:       req.Type,
		QuantityKg: quantity,
		OccurredAt: occurredAt,
		Origin:     req.Origin,
	}

	if err := s.movementRepo.Create(movement); err != nil {
		return nil, status.Errorf(codes.Internal, "erro ao registrar movimentação: %v", err)
	}

	return &inventoryv1.RegisterMovementResponse{Movement: toPBMovement(movement)}, nil
}

func (s *Server) GetStockByProduct(ctx context.Context, req *inventoryv1.GetStockByProductRequest) (*inventoryv1.GetStockByProductResponse, error) {
	if err := validateID(req.ProductId, "product_id"); err != nil {
		return nil, err
	}

	balances, err := s.lotRepo.ListByProductWithBalance(req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "erro ao consultar estoque: %v", err)
	}

	total := decimal.Zero
	pbLots := make([]*inventoryv1.LotStock, 0, len(balances))
	for _, b := range balances {
		total = total.Add(b.BalanceKg)
		pbLots = append(pbLots, &inventoryv1.LotStock{
			Lot:       toPBLot(&b.Lot),
			BalanceKg: b.BalanceKg.String(),
		})
	}

	return &inventoryv1.GetStockByProductResponse{
		TotalKg: total.String(),
		Lots:    pbLots,
	}, nil
}

func (s *Server) GetLotDetails(ctx context.Context, req *inventoryv1.GetLotDetailsRequest) (*inventoryv1.GetLotDetailsResponse, error) {
	if err := validateID(req.LotId, "lot_id"); err != nil {
		return nil, err
	}

	lot, err := s.findLot(req.LotId)
	if err != nil {
		return nil, err
	}

	movements, err := s.movementRepo.ListByLot(lot.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "erro ao listar movimentações: %v", err)
	}

	balance := decimal.Zero
	pbMovements := make([]*inventoryv1.StockMovement, 0, len(movements))
	for _, m := range movements {
		balance = balance.Add(m.QuantityKg)
		pbMovements = append(pbMovements, toPBMovement(&m))
	}

	return &inventoryv1.GetLotDetailsResponse{
		Lot:       toPBLot(lot),
		Movements: pbMovements,
		BalanceKg: balance.String(),
	}, nil
}

func (s *Server) findLot(id string) (*db.Lot, error) {
	lot, err := s.lotRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "lote %q não encontrado", id)
		}
		return nil, status.Errorf(codes.Internal, "erro ao buscar lote: %v", err)
	}
	return lot, nil
}

func validateID(id, field string) error {
	if _, err := uuid.Parse(id); err != nil {
		return status.Errorf(codes.InvalidArgument, "%s %q não é um UUID válido", field, id)
	}
	return nil
}

func toPBLot(lot *db.Lot) *inventoryv1.Lot {
	return &inventoryv1.Lot{
		Id:             lot.ID,
		ProductId:      lot.ProductID,
		PurchaseItemId: lot.PurchaseItemID,
		Safra:          lot.Safra,
		QuantityKg:     lot.QuantityKg.String(),
		ReceivedAt:     lot.ReceivedAt.Format("2006-01-02"),
	}
}

func toPBMovement(m *db.StockMovement) *inventoryv1.StockMovement {
	return &inventoryv1.StockMovement{
		Id:         m.ID,
		LotId:      m.LotID,
		Type:       m.Type,
		QuantityKg: m.QuantityKg.String(),
		OccurredAt: m.OccurredAt.Format(time.RFC3339),
		Origin:     m.Origin,
	}
}
