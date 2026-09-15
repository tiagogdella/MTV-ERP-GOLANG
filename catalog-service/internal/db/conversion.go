package db

import (
	"github.com/shopspring/decimal"
)

func (u *UnitOfMeasure) ToKg(quantity decimal.Decimal) decimal.Decimal {
	return quantity.Mul(u.ConversionFactorKg)
}

func (u *UnitOfMeasure) FromKg(kgAmount decimal.Decimal) decimal.Decimal {
	return kgAmount.Div(u.ConversionFactorKg)
}