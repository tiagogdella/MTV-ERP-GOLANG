package grpcserver

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"mtv-erp/catalog-service/internal/db"
	catalogv1 "mtv-erp/catalog-service/internal/pb/catalog/v1"
)

type Server struct {
	catalogv1.UnimplementedCatalogServiceServer
	productRepo  *db.ProductRepository
	unitRepo     *db.UnitOfMeasureRepository
	supplierRepo *db.SupplierRepository
}

func NewServer(productRepo *db.ProductRepository, unitRepo *db.UnitOfMeasureRepository, supplierRepo *db.SupplierRepository) *Server {
	return &Server{productRepo: productRepo, unitRepo: unitRepo, supplierRepo: supplierRepo}
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
	if err := validateID(req.Id); err != nil {
		return nil, err
	}

	if err := s.productRepo.Deactivate(req.Id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "produto %q não encontrado", req.Id)
		}
		return nil, err
	}

	return &catalogv1.DeactivateProductResponse{}, nil
}

func (s *Server) CreateUnitOfMeasure(ctx context.Context, req *catalogv1.CreateUnitOfMeasureRequest) (*catalogv1.CreateUnitOfMeasureResponse, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, status.Error(codes.InvalidArgument, "nome da unidade é obrigatório")
	}

	factor, err := decimal.NewFromString(req.ConversionFactorKg)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "fator de conversão %q não é um número válido", req.ConversionFactorKg)
	}
	if err := validateFactor(factor); err != nil {
		return nil, err
	}

	unit := &db.UnitOfMeasure{
		Name:               req.Name,
		ConversionFactorKg: factor,
	}

	if err := s.unitRepo.Create(unit); err != nil {
		return nil, status.Errorf(codes.Internal, "erro ao criar unidade: %v", err)
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
		return nil, status.Errorf(codes.Internal, "erro ao listar unidades: %v", err)
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
	if err := validateID(req.UnitId); err != nil {
		return nil, err
	}

	quantity, err := parsePositiveDecimal(req.Quantity, "quantidade")
	if err != nil {
		return nil, err
	}

	unit, err := s.findUnit(req.UnitId)
	if err != nil {
		return nil, err
	}

	return &catalogv1.ConvertToKgResponse{
		Kg: unit.ToKg(quantity).String(),
	}, nil
}

func (s *Server) ConvertFromKg(ctx context.Context, req *catalogv1.ConvertFromKgRequest) (*catalogv1.ConvertFromKgResponse, error) {
	if err := validateID(req.UnitId); err != nil {
		return nil, err
	}

	kg, err := parsePositiveDecimal(req.Kg, "kg")
	if err != nil {
		return nil, err
	}

	unit, err := s.findUnit(req.UnitId)
	if err != nil {
		return nil, err
	}

	return &catalogv1.ConvertFromKgResponse{
		Quantity: unit.FromKg(kg).String(),
	}, nil
}

func (s *Server) CreateSupplier(ctx context.Context, req *catalogv1.CreateSupplierRequest) (*catalogv1.CreateSupplierResponse, error) {
	supplier := &db.Supplier{
		Name:     req.Name,
		Document: req.Document,
		Address:  req.Address,
	}

	if err := s.supplierRepo.Create(supplier); err != nil {
		return nil, err
	}

	return &catalogv1.CreateSupplierResponse{
		Supplier: &catalogv1.Supplier{
			Id:       supplier.ID,
			Name:     supplier.Name,
			Document: supplier.Document,
			Address:  supplier.Address,
			Active:   supplier.Active,
		},
	}, nil
}

func (s *Server) ListSuppliers(ctx context.Context, req *catalogv1.ListSuppliersRequest) (*catalogv1.ListSuppliersResponse, error) {
	suppliers, err := s.supplierRepo.ListActive()
	if err != nil {
		return nil, err
	}

	var pbSuppliers []*catalogv1.Supplier
	for _, supplier := range suppliers {
		pbSuppliers = append(pbSuppliers, &catalogv1.Supplier{
			Id:       supplier.ID,
			Name:     supplier.Name,
			Document: supplier.Document,
			Address:  supplier.Address,
			Active:   supplier.Active,
		})
	}

	return &catalogv1.ListSuppliersResponse{
		Suppliers: pbSuppliers,
	}, nil
}

func (s *Server) DeactivateSupplier(ctx context.Context, req *catalogv1.DeactivateSupplierRequest) (*catalogv1.DeactivateSupplierResponse, error) {
	if err := validateID(req.Id); err != nil {
		return nil, err
	}

	if err := s.supplierRepo.Deactivate(req.Id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "fornecedor %q não encontrado", req.Id)
		}
		return nil, err
	}

	return &catalogv1.DeactivateSupplierResponse{}, nil
}

// Item 2.3
func (s *Server) GetProduct(ctx context.Context, req *catalogv1.GetProductRequest) (*catalogv1.GetProductResponse, error) {
	if err := validateID(req.Id); err != nil {
		return nil, err // Valida ID
	}

	product, err := s.productRepo.FindByID(req.Id)
	if err != nil {
		// Apenas o ErrRecordNotFound vira NotFound
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "produto %q não encontrado", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "erro ao buscar produto: %v", err)
	}

	return &catalogv1.GetProductResponse{
		Product: &catalogv1.Product{
			Id:     product.ID,
			Name:   product.Name,
			Active: product.Active,
		},
	}, nil
}

func (s *Server) GetSupplier(ctx context.Context, req *catalogv1.GetSupplierRequest) (*catalogv1.GetSupplierResponse, error) {
	if err := validateID(req.Id); err != nil {
		return nil, err
	}

	supplier, err := s.supplierRepo.FindByID(req.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "fornecedor %q não encontrado", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "erro ao buscar fornecedor: %v", err)
	}

	return &catalogv1.GetSupplierResponse{
		Supplier: &catalogv1.Supplier{
			Id:       supplier.ID,
			Name:     supplier.Name,
			Document: supplier.Document,
			Address:  supplier.Address,
			Active:   supplier.Active,
		},
	}, nil
}

func validateID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return status.Errorf(codes.InvalidArgument, "id %q não é um UUID válido", id)
	}
	return nil
}

var maxFactor = decimal.NewFromInt(100_000_000)

func validateFactor(f decimal.Decimal) error {
	if !f.IsPositive() {
		return status.Error(codes.InvalidArgument, "fator de conversão deve ser maior que zero")
	}
	if !f.Equal(f.Round(4)) {
		return status.Error(codes.InvalidArgument, "fator de conversão aceita no máximo 4 casas decimais")
	}
	if f.GreaterThanOrEqual(maxFactor) {
		return status.Error(codes.InvalidArgument, "fator de conversão deve ser menor que 100.000.000")
	}
	return nil
}

func parsePositiveDecimal(value, field string) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Decimal{}, status.Errorf(codes.InvalidArgument, "%s %q não é um número válido", field, value)
	}
	if !d.IsPositive() {
		return decimal.Decimal{}, status.Errorf(codes.InvalidArgument, "%s deve ser maior que zero", field)
	}
	return d, nil
}

func (s *Server) findUnit(id string) (*db.UnitOfMeasure, error) {
	unit, err := s.unitRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "unidade %q não encontrada", id)
		}
		return nil, status.Errorf(codes.Internal, "erro ao buscar unidade: %v", err)
	}
	return unit, nil
}