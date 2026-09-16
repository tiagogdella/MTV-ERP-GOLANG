package grpcserver

import (
	"context"

	"mtv-erp/catalog-service/internal/db"
	catalogv1 "mtv-erp/catalog-service/internal/pb/catalog/v1"
	"github.com/shopspring/decimal"
)

type Server struct {
	catalogv1.UnimplementedCatalogServiceServer
	productRepo *db.ProductRepository
	unitRepo    *db.UnitOfMeasureRepository
}

func NewServer(productRepo *db.ProductRepository, unitRepo *db.UnitOfMeasureRepository) *Server {
	return &Server{productRepo: productRepo, unitRepo: unitRepo}
}

func (s *Server) CreateProduct(ctx context.Context, req *catalogv1.CreateProductRequest) (*catalogv1.CreateProductResponse, error) {
	product := &db.Product{
		Name: req.Name,
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, err
	}

	return &catalogv1.CreateProductResponse{
		Product: &catalogv1.Product{
			Id:     product.ID,
			Name:   product.Name,
			Active: product.Active,
		},
	}, nil
}

func (s *Server) ListProducts(ctx context.Context, req *catalogv1.ListProductsRequest) (*catalogv1.ListProductsResponse, error) {
	products, err := s.productRepo.ListActive()
	if err != nil {
		return nil, err
	}

	var pbProducts []*catalogv1.Product
	for _, product := range products {
		pbProducts = append(pbProducts, &catalogv1.Product{
			Id:     product.ID,
			Name:   product.Name,
			Active: product.Active,
		})
	}

	return &catalogv1.ListProductsResponse{
		Products: pbProducts,
	}, nil
}

func (s *Server) DeactivateProduct(ctx context.Context, req *catalogv1.DeactivateProductRequest) (*catalogv1.DeactivateProductResponse, error) {
	if err := s.productRepo.Deactivate(req.Id); err != nil {
		return nil, err
	}

	return &catalogv1.DeactivateProductResponse{}, nil
}

func (s *Server) CreateUnitOfMeasure(ctx context.Context, req *catalogv1.CreateUnitOfMeasureRequest) (*catalogv1.CreateUnitOfMeasureResponse, error) {
	factor, err := decimal.NewFromString(req.ConversionFactorKg)
	if err != nil {
		return nil, err
	}

	unit := &db.UnitOfMeasure{
		Name:               req.Name,
		ConversionFactorKg: factor,
	}

	if err := s.unitRepo.Create(unit); err != nil {
		return nil, err
	}

	return &catalogv1.CreateUnitOfMeasureResponse{
		UnitOfMeasure: &catalogv1.UnitOfMeasure{
			Id:                 unit.ID,
			Name:               unit.Name,
			ConversionFactorKg: unit.ConversionFactorKg.String(),
		},
	}, nil
}

func (s *Server) ListUnitsOfMeasure(ctx context.Context, req *catalogv1.ListUnitsOfMeasureRequest) (*catalogv1.ListUnitsOfMeasureResponse, error) {
	units, err := s.unitRepo.List()
	if err != nil {
		return nil, err
	}

	var pbUnits []*catalogv1.UnitOfMeasure
	for _, unit := range units {
		pbUnits = append(pbUnits, &catalogv1.UnitOfMeasure{
			Id:                 unit.ID,
			Name:               unit.Name,
			ConversionFactorKg: unit.ConversionFactorKg.String(),
		})
	}

	return &catalogv1.ListUnitsOfMeasureResponse{
		Units: pbUnits,
	}, nil
}

func (s *Server) ConvertToKg(ctx context.Context, req *catalogv1.ConvertToKgRequest) (*catalogv1.ConvertToKgResponse, error) {
	unit, err := s.unitRepo.FindByID(req.UnitId)
	if err != nil {
		return nil, err
	}

	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		return nil, err
	}

	kg := unit.ToKg(quantity)

	return &catalogv1.ConvertToKgResponse{
		Kg: kg.String(),
	}, nil
}

func (s *Server) ConvertFromKg(ctx context.Context, req *catalogv1.ConvertFromKgRequest) (*catalogv1.ConvertFromKgResponse, error) {
	unit, err := s.unitRepo.FindByID(req.UnitId)
	if err != nil {
		return nil, err
	}

	kg, err := decimal.NewFromString(req.Kg)
	if err != nil {
		return nil, err
	}

	quantity := unit.FromKg(kg)

	return &catalogv1.ConvertFromKgResponse{
		Quantity: quantity.String(),
	}, nil
}
