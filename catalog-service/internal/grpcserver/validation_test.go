package grpcserver

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidateFactor(t *testing.T) {
	tests := []struct {
		name     string
		factor   string
		wantCode codes.Code
	}{
		{"inteiro", "30", codes.OK},
		{"4 casas decimais", "25.1234", codes.OK},
		{"zeros à direita não contam", "25.10000", codes.OK},
		{"logo abaixo do máximo", "99999999.9999", codes.OK},
		{"zero", "0", codes.InvalidArgument},
		{"negativo", "-25", codes.InvalidArgument},
		{"5 casas decimais", "25.12345", codes.InvalidArgument},
		{"no máximo", "100000000", codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFactor(decimal.RequireFromString(tt.factor))
			assert.Equal(t, tt.wantCode, status.Code(err))
		})
	}
}
