package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"mtv-erp/purchasing-service/internal/db"
	catalogv1 "mtv-erp/purchasing-service/internal/pb/catalog/v1"
	inventoryv1 "mtv-erp/purchasing-service/internal/pb/inventory/v1"
)

type PurchaseService struct {
	purchaseRepo    *db.PurchaseRepository
	catalogClient   catalogv1.CatalogServiceClient
	inventoryClient inventoryv1.InventoryServiceClient
}

func NewPurchaseService(purchaseRepo *db.PurchaseRepository, catalogClient catalogv1.CatalogServiceClient, inventoryClient inventoryv1.InventoryServiceClient) *PurchaseService {
	return &PurchaseService{
		purchaseRepo:    purchaseRepo,
		catalogClient:   catalogClient,
		inventoryClient: inventoryClient,
	}
}

type CreatePurchaseItemInput struct {
	ProductID string
	UnitID    string
	Safra     string
	Quantity  decimal.Decimal
}

type CreatePurchaseInput struct {
	SupplierID    string
	InvoiceNumber string
	InvoiceDate   time.Time
	InvoiceValue  decimal.Decimal
	Items         []CreatePurchaseItemInput
}

type CreatePurchaseItemResult struct {
	Item  db.PurchaseItem
	LotID string
}

type CreatePurchaseResult struct {
	Purchase db.Purchase
	Items    []CreatePurchaseItemResult
}

func (s *PurchaseService) CreatePurchase(ctx context.Context, input CreatePurchaseInput) (*CreatePurchaseResult, error) {
	if err := validatePurchaseInput(input); err != nil {
		return nil, err
	}

	if _, err := s.catalogClient.GetSupplier(ctx, &catalogv1.GetSupplierRequest{Id: input.SupplierID}); err != nil {
		return nil, translateCatalogError(err, "")
	}

	purchase := &db.Purchase{
		SupplierID:    input.SupplierID,
		InvoiceNumber: input.InvoiceNumber,
		InvoiceDate:   input.InvoiceDate,
		InvoiceValue:  input.InvoiceValue,
	}

	items := make([]db.PurchaseItem, 0, len(input.Items))
	kgByItem := make([]decimal.Decimal, 0, len(input.Items))

	for i, itemInput := range input.Items {
		prefix := fmt.Sprintf("item %d: ", i+1)

		if _, err := s.catalogClient.GetProduct(ctx, &catalogv1.GetProductRequest{Id: itemInput.ProductID}); err != nil {
			return nil, translateCatalogError(err, prefix)
		}

		convResp, err := s.catalogClient.ConvertToKg(ctx, &catalogv1.ConvertToKgRequest{
			UnitId:   itemInput.UnitID,
			Quantity: itemInput.Quantity.String(),
		})
		if err != nil {
			return nil, translateCatalogError(err, prefix)
		}

		kg, err := decimal.NewFromString(convResp.Kg)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "%skg convertido inválido: %v", prefix, err)
		}

		items = append(items, db.PurchaseItem{
			ProductID: itemInput.ProductID,
			UnitID:    itemInput.UnitID,
			Safra:     itemInput.Safra,
			Quantity:  itemInput.Quantity,
		})
		kgByItem = append(kgByItem, kg)
	}

	if err := s.purchaseRepo.Create(purchase, items); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, status.Errorf(codes.AlreadyExists, "a nota fiscal %q já foi lançada para este fornecedor", input.InvoiceNumber)
		}
		return nil, status.Errorf(codes.Internal, "falha ao salvar compra: %v", err)
	}

	result := &CreatePurchaseResult{
		Purchase: *purchase,
		Items:    make([]CreatePurchaseItemResult, 0, len(items)),
	}

	for i, item := range items {
		lotResp, err := s.inventoryClient.CreateLot(ctx, &inventoryv1.CreateLotRequest{
			ProductId:      item.ProductID,
			PurchaseItemId: item.ID,
			Safra:          item.Safra,
			QuantityKg:     kgByItem[i].String(),
			ReceivedAt:     input.InvoiceDate.Format("2006-01-02"),
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "compra %s gravada, mas falhou ao criar o lote do item %d: %v", purchase.ID, i+1, status.Convert(err).Message())
		}

		result.Items = append(result.Items, CreatePurchaseItemResult{
			Item:  item,
			LotID: lotResp.Lot.Id,
		})
	}

	return result, nil
}

func validatePurchaseInput(in CreatePurchaseInput) error {
	if strings.TrimSpace(in.InvoiceNumber) == "" {
		return status.Error(codes.InvalidArgument, "número da nota fiscal é obrigatório")
	}
	if !in.InvoiceValue.IsPositive() {
		return status.Error(codes.InvalidArgument, "valor da nota fiscal deve ser maior que zero")
	}
	if len(in.Items) == 0 {
		return status.Error(codes.InvalidArgument, "a compra precisa de pelo menos um item")
	}
	for i, item := range in.Items {
		if strings.TrimSpace(item.Safra) == "" {
			return status.Errorf(codes.InvalidArgument, "item %d: safra é obrigatória", i+1)
		}
		if !item.Quantity.IsPositive() {
			return status.Errorf(codes.InvalidArgument, "item %d: quantidade deve ser maior que zero", i+1)
		}
	}
	return nil
}

func translateCatalogError(err error, prefix string) error {
	st := status.Convert(err)
	switch st.Code() {
	case codes.NotFound:
		return status.Error(codes.FailedPrecondition, prefix+st.Message())
	case codes.InvalidArgument:
		return status.Error(codes.InvalidArgument, prefix+st.Message())
	case codes.Unavailable:
		return status.Error(codes.Unavailable, "catálogo indisponível no momento")
	case codes.DeadlineExceeded:
		return status.Error(codes.DeadlineExceeded, "catálogo não respondeu a tempo")
	default:
		return status.Errorf(codes.Internal, "%serro ao consultar o catálogo: %s", prefix, st.Message())
	}
}
