package productHandler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"products/internal/db/product"
	"testing"
	"time"
)

type mockProductQuerier struct {
	createProductFunc         func(ctx context.Context, arg product.CreateProductParams) error
	deleteProductFunc         func(ctx context.Context, id int32) error
	getProductByIdFunc        func(ctx context.Context, id int32) (product.Product, error)
	getProductsByPlatformFunc func(ctx context.Context, platformID int32) ([]product.Product, error)
	updateProductFunc         func(ctx context.Context, arg product.UpdateProductParams) error
}

func (m *mockProductQuerier) CreateProduct(ctx context.Context, arg product.CreateProductParams) error {
	return m.createProductFunc(ctx, arg)
}

func (m *mockProductQuerier) DeleteProduct(ctx context.Context, id int32) error {
	return m.deleteProductFunc(ctx, id)
}

func (m *mockProductQuerier) GetProductById(ctx context.Context, id int32) (product.Product, error) {
	return m.getProductByIdFunc(ctx, id)
}

func (m *mockProductQuerier) GetProductsByPlatform(ctx context.Context, platformID int32) ([]product.Product, error) {
	return m.getProductsByPlatformFunc(ctx, platformID)
}

func (m *mockProductQuerier) UpdateProduct(ctx context.Context, arg product.UpdateProductParams) error {
	return m.updateProductFunc(ctx, arg)
}

func TestCreateProduct(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		mockSetup      func(m *mockProductQuerier)
		expectedStatus int
	}{
		{
			name: "Success",
			requestBody: CreateProductRequest{
				PlatformID: 1,
				Name:       "Test Product",
			},
			mockSetup: func(m *mockProductQuerier) {
				m.createProductFunc = func(ctx context.Context, arg product.CreateProductParams) error {
					if arg.Name != "Test Product" {
						return errors.New("unexpected name")
					}
					if arg.PlatformID != 1 {
						return errors.New("unexpected platform id")
					}
					return nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *mockProductQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "DB Failure",
			requestBody: CreateProductRequest{
				PlatformID: 1,
				Name:       "Fail Product",
			},
			mockSetup: func(m *mockProductQuerier) {
				m.createProductFunc = func(ctx context.Context, arg product.CreateProductParams) error {
					return errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Missing Name",
			requestBody: CreateProductRequest{
				PlatformID: 1,
			},
			mockSetup:      func(m *mockProductQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing PlatformID",
			requestBody: CreateProductRequest{
				Name: "Test Product",
			},
			mockSetup:      func(m *mockProductQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockProductQuerier{}
			tt.mockSetup(mock)
			h := NewProductHandler(mock)

			var body []byte
			if s, ok := tt.requestBody.(string); ok {
				body = []byte(s)
			} else {
				var err error
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("json.Marshal requestBody failed: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			h.CreateProduct(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestDeleteProduct(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(m *mockProductQuerier)
		expectedStatus int
	}{
		{
			name: "Success",
			id:   "1",
			mockSetup: func(m *mockProductQuerier) {
				m.deleteProductFunc = func(ctx context.Context, id int32) error {
					if id != 1 {
						return errors.New("unexpected id")
					}
					return nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Invalid ID (Not numeric)",
			id:             "abc",
			mockSetup:      func(m *mockProductQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid ID (Zero)",
			id:             "0",
			mockSetup:      func(m *mockProductQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid ID (Negative)",
			id:             "-1",
			mockSetup:      func(m *mockProductQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "DB Failure",
			id:   "1",
			mockSetup: func(m *mockProductQuerier) {
				m.deleteProductFunc = func(ctx context.Context, id int32) error {
					return errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Timeout context set to ~5s",
			id:   "1",
			mockSetup: func(m *mockProductQuerier) {
				m.deleteProductFunc = func(ctx context.Context, id int32) error {
					deadline, ok := ctx.Deadline()
					if !ok {
						return errors.New("deadline not set")
					}
					diff := time.Until(deadline)
					if diff < 4900*time.Millisecond || diff > 5100*time.Millisecond {
						return errors.New("deadline not approximately 5s")
					}
					return nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "Idempotent non-existent delete",
			id:   "999",
			mockSetup: func(m *mockProductQuerier) {
				m.deleteProductFunc = func(ctx context.Context, id int32) error {
					return sql.ErrNoRows
				}
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockProductQuerier{}
			tt.mockSetup(mock)
			h := NewProductHandler(mock)

			req := httptest.NewRequest(http.MethodDelete, "/api/products/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			rr := httptest.NewRecorder()

			h.DeleteProduct(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}
		})
	}
}
