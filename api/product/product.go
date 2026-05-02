package productHandler

import "products/internal/db/product"

type ProductHandler struct {
	queries product.Querier
}

func NewProductHandler(q product.Querier) *ProductHandler {
	return &ProductHandler{
		queries: q,
	}
}
