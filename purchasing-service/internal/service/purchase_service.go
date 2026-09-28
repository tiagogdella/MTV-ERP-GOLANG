package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
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
	//Validamos se fornecedor existe
	if _, err := s.catalogClient.GetSupplier(ctx, &catalogv1.GetSupplierRequest{Id: input.SupplierID}); err != nil {
		return nil, fmt.Errorf("fornecedor inválido: %w", err)
	}

	purchase := &db.Purchase{
		SupplierID:    input.SupplierID,
		InvoiceNumber: input.InvoiceNumber,
		InvoiceDate:   input.InvoiceDate,
		InvoiceValue:  input.InvoiceValue,
	}

	items := make([]db.PurchaseItem, 0, len(input.Items))
	kgByItem := make([]decimal.Decimal, 0, len(input.Items))

	for _, itemInput := range input.Items {
		//Validamos se produto existe
		if _, err := s.catalogClient.GetProduct(ctx, &catalogv1.GetProductRequest{Id: itemInput.ProductID}); err != nil {
			return nil, fmt.Errorf("produto inválido: %w", err)
		}

		convResp, err := s.catalogClient.ConvertToKg(ctx, &catalogv1.ConvertToKgRequest{
			UnitId:   itemInput.UnitID,
			Quantity: itemInput.Quantity.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("unidade inválida: %w", err)
		}

		kg, err := decimal.NewFromString(convResp.Kg)
		if err != nil {
			return nil, fmt.Errorf("kg convertido inválido: %w", err)
		}

		items = append(items, db.PurchaseItem{
			ProductID: itemInput.ProductID,
			UnitID:    itemInput.UnitID,
			Safra:     itemInput.Safra,
			Quantity:  itemInput.Quantity,
		})
		kgByItem = append(kgByItem, kg)
	}

	//Criamos purchase para pegar o ID que vai ser usado em purchase_item
	if err := s.purchaseRepo.Create(purchase, items); err != nil {
		return nil, fmt.Errorf("falha ao salvar compra: %w", err)
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
			return nil, fmt.Errorf("falha ao criar lote pro item %s: %w", item.ID, err)
		}

		result.Items = append(result.Items, CreatePurchaseItemResult{
			Item:  item,
			LotID: lotResp.Lot.Id,
		})
	}

	return result, nil
}
