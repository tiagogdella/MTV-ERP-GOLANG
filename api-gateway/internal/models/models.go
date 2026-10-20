package models

// Catalog
type Product struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Active bool  `json:"active"`
}

type ProductInput struct {
	Name string `json:"name"`
}

type ProductPatchInput struct {
	Active *bool `json:"active"`
}

type Supplier struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Document string `json:"document"`
	Address  string `json:"address"`
	Active   bool   `json:"active"`
}

type SupplierInput struct {
	Name     string `json:"name"`
	Document string `json:"document"`
	Address  string `json:"address"`
}

type SupplierPatchInput struct {
	Active *bool `json:"active"`
}

type Address struct {
	CEP          string `json:"cep"`
	Street       string `json:"street"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
}

// Auth
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginOutput struct {
	Token string `json:"token"`
}

type UserInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Units
type Unit struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	ConversionFactorKg string `json:"conversionFactorKg"`
}

type UnitInput struct {
	Name               string `json:"name"`
	ConversionFactorKg string `json:"conversionFactorKg"`
}

type Conversion struct {
	UnitID   string `json:"unitId"`
	Quantity string `json:"quantity"`
	Kg       string `json:"kg"`
}
