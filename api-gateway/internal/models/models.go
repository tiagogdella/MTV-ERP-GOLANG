package models

type Product struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Active bool   `json:"active"`
}
