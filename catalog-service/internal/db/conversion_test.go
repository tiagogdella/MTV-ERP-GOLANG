package db

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestToKg_CasoNormal(t *testing.T) {
	unit := UnitOfMeasure{ConversionFactorKg: decimal.NewFromFloat(30.0)} // fardo 30kg

	result := unit.ToKg(decimal.NewFromFloat(1)) // 1 fardo

	expected := decimal.NewFromFloat(30.0) // conta de cabeça: 1 × 30 = 30
	if !result.Equal(expected) {
		t.Fatalf("esperado %s, recebido %s", expected, result)
	}
}

func TestToKg_fracao(t *testing.T) {
	unit := UnitOfMeasure{ConversionFactorKg: decimal.NewFromFloat(25.49)}

	result := unit.ToKg(decimal.NewFromFloat(2.5)) 

	expected :=  decimal.NewFromFloat(63.725)
	if !result.Equal(expected) {
		t.Fatalf("esperado %s, recebido %s", expected, result)
	}
}

func TestToKg_fromKg(t *testing.T) {
	unit := UnitOfMeasure{ConversionFactorKg: decimal.NewFromFloat(5.00)}

	result1 := unit.ToKg(decimal.NewFromFloat(158.57))
	result2 := unit.FromKg(result1)
	

	expected :=  decimal.NewFromFloat(158.57)
	if !result2.Equal(expected) {
		t.Fatalf("esperado %s, recebido %s", expected, result2)
	}
}
