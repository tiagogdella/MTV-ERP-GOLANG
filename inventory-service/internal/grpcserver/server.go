package grpcserver

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	quantity, err := decimal.NewFromString(req.QuantityKg)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "quantity_kg inválido: %v", err)
	}

	receivedAt, err := time.Parse("2006-01-02", req.ReceivedAt)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "received_at inválido, use AAAA-MM-DD: %v", err)
	}

	lot := &db.Lot{
		ProductID:      req.ProductId,
		PurchaseItemID: req.PurchaseItemId,
		Safra:          req.Safra,
		QuantityKg:     quantity,
		ReceivedAt:     receivedAt,
	}

	if err := s.lotRepo.Create(lot); err != nil {
		return nil, err
	}

	return &inventoryv1.CreateLotResponse{Lot: toPBLot(lot)}, nil
}

func (s *Server) RegisterMovement(ctx context.Context, req *inventoryv1.RegisterMovementRequest) (*inventoryv1.RegisterMovementResponse, error) {
	// validação explícita: sem lote não existe estoque (não é só a FK do banco)
	if _, err := s.lotRepo.FindByID(req.LotId); err != nil {
		return nil, status.Errorf(codes.NotFound, "lote %q não encontrado", req.LotId)
	}

	quantity, err := decimal.NewFromString(req.QuantityKg)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "quantity_kg inválido: %v", err)
	}

	occurredAt, err := time.Parse(time.RFC3339, req.OccurredAt)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "occurred_at inválido, use RFC3339: %v", err)
	}

	movement := &db.StockMovement{
		LotID:      req.LotId,
		Type:       req.Type,
		QuantityKg: quantity,
		OccurredAt: occurredAt,
		Origin:     req.Origin,
	}

	if err := s.movementRepo.Create(movement); err != nil {
		return nil, err
	}

	return &inventoryv1.RegisterMovementResponse{Movement: toPBMovement(movement)}, nil
}

func (s *Server) GetStockByProduct(ctx context.Context, req *inventoryv1.GetStockByProductRequest) (*inventoryv1.GetStockByProductResponse, error) {
	lots, err := s.lotRepo.ListByProduct(req.ProductId)
	if err != nil {
		return nil, err
	}

	total := decimal.Zero
	for _, lot := range lots {
		movements, err := s.movementRepo.ListByLot(lot.ID)
		if err != nil {
			return nil, err
		}
		for _, m := range movements {
			total = total.Add(m.QuantityKg)
		}
	}

	return &inventoryv1.GetStockByProductResponse{TotalKg: total.String()}, nil
}

func (s *Server) GetLotDetails(ctx context.Context, req *inventoryv1.GetLotDetailsRequest) (*inventoryv1.GetLotDetailsResponse, error) {
	lot, err := s.lotRepo.FindByID(req.LotId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "lote %q não encontrado", req.LotId)
	}

	movements, err := s.movementRepo.ListByLot(lot.ID)
	if err != nil {
		return nil, err
	}

	balance := decimal.Zero
	var pbMovements []*inventoryv1.StockMovement
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
