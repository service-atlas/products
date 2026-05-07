package product

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"products/internal"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func NewProductHandler(db DBTX) Handler {
	queries := &Queries{
		db: db,
	}
	return &productHandler{
		queries: queries,
	}
}

type Handler interface {
	CreateProduct(w http.ResponseWriter, r *http.Request)
	GetProductsByPlatform(w http.ResponseWriter, r *http.Request)
	GetProductById(w http.ResponseWriter, r *http.Request)
	UpdateProduct(w http.ResponseWriter, r *http.Request)
	DeleteProduct(w http.ResponseWriter, r *http.Request)
}

type createProductRequest struct {
	Name        string `json:"name"`
	PlatformID  int32  `json:"platform_id"`
	Description string `json:"description"`
}

type updateProductRequest struct {
	PlatformID  int32  `json:"platform_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type productHandler struct {
	queries Querier
}

func (h *productHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.PlatformID == 0 {
		http.Error(w, "Name and platform ID are required", http.StatusBadRequest)
		return
	}

	params := CreateProductParams{
		Name:       req.Name,
		PlatformID: req.PlatformID,
		Description: pgtype.Text{
			Valid:  req.Description != "",
			String: req.Description,
		},
		Timestamp: pgtype.Timestamptz{
			Valid: true,
			Time:  time.Now().UTC(),
		},
	}

	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := h.queries.CreateProduct(contextWithTimeOut, params); err != nil {
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *productHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if _, err := h.queries.DeleteProduct(contextWithTimeOut, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *productHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req updateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.PlatformID == 0 {
		http.Error(w, "Name and platform ID are required", http.StatusBadRequest)
		return
	}

	params := UpdateProductParams{
		ID:         id,
		PlatformID: req.PlatformID,
		Name:       req.Name,
		Description: pgtype.Text{
			Valid:  req.Description != "",
			String: req.Description,
		},
		UpdatedAt: pgtype.Timestamptz{
			Valid: true,
			Time:  time.Now().UTC(),
		},
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if _, err := h.queries.UpdateProduct(ctx, params); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetProductsByPlatform fetches products by platform ID.
func (h *productHandler) GetProductsByPlatform(w http.ResponseWriter, r *http.Request) {
	platformID, ok := internal.GetIntFromRequestPath("platform_id", r)
	if !ok {
		http.Error(w, "Invalid platform ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	products, err := h.queries.GetProductsByPlatform(ctx, platformID)
	if err != nil {
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}

	if products == nil {
		products = []Product{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(products); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetProductById fetches a single product by ID.
func (h *productHandler) GetProductById(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	product, err := h.queries.GetProductById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(product); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
