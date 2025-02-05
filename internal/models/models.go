package models

type Result struct {
	TotalItems      int     `json:"total_items"`
	TotalCategories int     `json:"total_categories"`
	TotalPrice      float64 `json:"total_price"`
}

type Product struct {
	Id           int
	CreationTime string
	Name         string
	Category     string
	Price        float64
}
