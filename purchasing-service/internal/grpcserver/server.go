package grpcserver

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	purchasingv1 "mtv-erp/purchasing-service/internal/pb/purchasing/v1"
	"mtv-erp/purchasing-service/internal/service"
)

type Server struct {
	purchasingv1.UnimplementedPurchasingServiceServer
	purchaseService *service.PurchaseService
}

func NewServer(purchaseService *service.PurchaseService) *Server {
	return &Server{purchaseService: purchaseService}
}

func (s *Server) CreatePurchase(ctx context.Context, req *purchasingv1.CreatePurchaseRequest) (*purchasingv1.CreatePurchaseResponse, error) {
	invoiceDate, err := time.Parse("2006-01-02", req.InvoiceDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invoice_date inválido, use AAAA-MM-DD: %v", err)
	}

	invoiceValue, err := decimal.NewFromString(req.InvoiceValue)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invoice_value inválido: %v", err)
	}

	items := make([]service.CreatePurchaseItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		quantity, err := decimal.NewFromString(item.Quantity)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "quantity inválida no item: %v", err)
		}

		items = append(items, service.CreatePurchaseItemInput{
			ProductID: item.ProductId,
			UnitID:    item.UnitId,
			Safra:     item.Safra,
			Quantity:  quantity,
		})
	}

	result, err := s.purchaseService.CreatePurchase(ctx, service.CreatePurchaseInput{
		SupplierID:    req.SupplierId,
		InvoiceNumber: req.InvoiceNumber,
		InvoiceDate:   invoiceDate,
		InvoiceValue:  invoiceValue,
		Items:         items,
	})
	if err != nil {
		return nil, err
	}

	pbItems := make([]*purchasingv1.PurchaseItem, 0, len(result.Items))
	for _, item := range result.Items {
		pbItems = append(pbItems, &purchasingv1.PurchaseItem{
			Id:        item.Item.ID,
			ProductId: item.Item.ProductID,
			UnitId:    item.Item.UnitID,
			Safra:     item.Item.Safra,
			Quantity:  item.Item.Quantity.String(),
			LotId:     item.LotID,
		})
	}

	return &purchasingv1.CreatePurchaseResponse{
		Purchase: &purchasingv1.Purchase{
			Id:            result.Purchase.ID,
			SupplierId:    result.Purchase.SupplierID,
			InvoiceNumber: result.Purchase.InvoiceNumber,
			InvoiceDate:   result.Purchase.InvoiceDate.Format("2006-01-02"),
			InvoiceValue:  result.Purchase.InvoiceValue.String(),
			Items:         pbItems,
		},
	}, nil
}
