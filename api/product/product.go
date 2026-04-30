package productHandler

import "products/internal/db/product"

type ProductHandler struct {
	queries product.Querier
}
