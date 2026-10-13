package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"mtv-erp/api-gateway/internal/models"
)

func TestValidateSupplierInput(t *testing.T) {
	tests := []struct {
		name   string
		input  models.SupplierInput
		wantOK bool
	}{
		{"todos os campos preenchidos", models.SupplierInput{Name: "Fornecedor", Document: "123", Address: "Rua 1"}, true},
		{"name vazio", models.SupplierInput{Document: "123", Address: "Rua 1"}, false},
		{"name só com espaços", models.SupplierInput{Name: "   ", Document: "123", Address: "Rua 1"}, false},
		{"document vazio", models.SupplierInput{Name: "Fornecedor", Address: "Rua 1"}, false},
		{"address vazio", models.SupplierInput{Name: "Fornecedor", Document: "123"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := validateSupplierInput(tt.input)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestValidateProductPatchInput(t *testing.T) {
	falseValue := false
	trueValue := true

	tests := []struct {
		name   string
		input  models.ProductPatchInput
		wantOK bool
	}{
		{"active ausente", models.ProductPatchInput{Active: nil}, false},
		{"active true (reativar não existe)", models.ProductPatchInput{Active: &trueValue}, false},
		{"active false", models.ProductPatchInput{Active: &falseValue}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := validateProductPatchInput(tt.input)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}
