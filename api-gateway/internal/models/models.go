package models

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