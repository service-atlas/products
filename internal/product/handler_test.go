package product

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"products/internal"
	"products/internal/product/db"
	"testing"
	"time"
)

type mockProductService struct {
	createProductFunc         func(ctx context.Context, req createProductRequest) (db.Product, error)
	getProductsByPlatformFunc func(ctx context.Context, platformID int) ([]db.Product, error)
	getProductByIdFunc        func(ctx context.Context, id int) (db.Product, error)
	updateProductFunc         func(ctx context.Context, req updateProductRequest, id int) (int, error)
	deleteProductFunc         func(ctx context.Context, id int) (int, error)
}

func (m *mockProductService) CreateProduct(ctx context.Context, req createProductRequest) (db.Product, error) {
	return m.createProductFunc(ctx, req)
}

func (m *mockProductService) GetProductsByPlatform(ctx context.Context, platformID int) ([]db.Product, error) {
	return m.getProductsByPlatformFunc(ctx, platformID)
}

func (m *mockProductService) GetProductById(ctx context.Context, id int) (db.Product, error) {
	return m.getProductByIdFunc(ctx, id)
}

func (m *mockProductService) UpdateProduct(ctx context.Context, req updateProductRequest, id int) (int, error) {
	return m.updateProductFunc(ctx, req, id)
}

func (m *mockProductService) DeleteProduct(ctx context.Context, id int) (int, error) {
	return m.deleteProductFunc(ctx, id)
}

func TestCreateProduct(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		mockSetup      func(m *mockProductService)
		expectedStatus int
	}{
		{
			name: "Success",
			requestBody: createProductRequest{
				PlatformID: 1,
				Name:       "Test Product",
			},
			mockSetup: func(m *mockProductService) {
				m.createProductFunc = func(ctx context.Context, req createProductRequest) (db.Product, error) {
					if req.Name != "Test Product" {
						return db.Product{}, errors.New("unexpected name")
					}
					if req.PlatformID != 1 {
						return db.Product{}, errors.New("unexpected platform id")
					}
					return db.Product{ID: 1, Name: req.Name, PlatformID: req.PlatformID}, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "DB Failure",
			requestBody: createProductRequest{
				PlatformID: 1,
				Name:       "Fail Product",
			},
			mockSetup: func(m *mockProductService) {
				m.createProductFunc = func(ctx context.Context, req createProductRequest) (db.Product, error) {
					return db.Product{}, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Missing Name",
			requestBody: createProductRequest{
				PlatformID: 1,
			},
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing PlatformID",
			requestBody: createProductRequest{
				Name: "Test Product",
			},
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Whitespace Name",
			requestBody: createProductRequest{
				PlatformID: 1,
				Name:       "   ",
			},
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockProductService{}
			tt.mockSetup(mock)
			h := &handler{service: mock}

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
		mockSetup      func(m *mockProductService)
		expectedStatus int
	}{
		{
			name: "Success",
			id:   "1",
			mockSetup: func(m *mockProductService) {
				m.deleteProductFunc = func(ctx context.Context, id int) (int, error) {
					if id != 1 {
						return 0, errors.New("unexpected id")
					}
					return 1, nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Invalid ID (Not numeric)",
			id:             "abc",
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid ID (Zero)",
			id:             "0",
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid ID (Negative)",
			id:             "-1",
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "DB Failure",
			id:   "1",
			mockSetup: func(m *mockProductService) {
				m.deleteProductFunc = func(ctx context.Context, id int) (int, error) {
					return 0, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Timeout context set to ~5s",
			id:   "1",
			mockSetup: func(m *mockProductService) {
				m.deleteProductFunc = func(ctx context.Context, id int) (int, error) {
					deadline, ok := ctx.Deadline()
					if !ok {
						return 0, errors.New("deadline not set")
					}
					diff := time.Until(deadline)
					if diff < 4900*time.Millisecond || diff > 5100*time.Millisecond {
						return 0, errors.New("deadline not approximately 5s")
					}
					return 1, nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "Not Found",
			id:   "999",
			mockSetup: func(m *mockProductService) {
				m.deleteProductFunc = func(ctx context.Context, id int) (int, error) {
					return 0, internal.NewNotFoundError(id, "Product")
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockProductService{}
			tt.mockSetup(mock)
			h := &handler{service: mock}

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

func TestGetProductsByPlatform(t *testing.T) {
	tests := []struct {
		name           string
		platformID     string
		mockSetup      func(m *mockProductService)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:       "Success",
			platformID: "1",
			mockSetup: func(m *mockProductService) {
				m.getProductsByPlatformFunc = func(ctx context.Context, platformID int) ([]db.Product, error) {
					if platformID != 1 {
						return nil, errors.New("unexpected platform id")
					}
					return []db.Product{
						{ID: 1, PlatformID: 1, Name: "Product 1"},
						{ID: 2, PlatformID: 1, Name: "Product 2"},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:       "Empty list (Nil guard)",
			platformID: "1",
			mockSetup: func(m *mockProductService) {
				m.getProductsByPlatformFunc = func(ctx context.Context, platformID int) ([]db.Product, error) {
					return nil, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "Invalid Platform ID",
			platformID:     "abc",
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "DB Failure",
			platformID: "1",
			mockSetup: func(m *mockProductService) {
				m.getProductsByPlatformFunc = func(ctx context.Context, platformID int) ([]db.Product, error) {
					return nil, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockProductService{}
			tt.mockSetup(mock)
			h := &handler{service: mock}

			req := httptest.NewRequest(http.MethodGet, "/api/platforms/"+tt.platformID+"/products", nil)
			req.SetPathValue("id", tt.platformID)
			rr := httptest.NewRecorder()

			h.GetProductsByPlatform(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var products []db.Product
				if err := json.NewDecoder(rr.Body).Decode(&products); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if len(products) != tt.expectedCount {
					t.Errorf("expected %v products, got %v", tt.expectedCount, len(products))
				}
				if rr.Header().Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %v", rr.Header().Get("Content-Type"))
				}
			}
		})
	}
}

func TestGetProductById(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(m *mockProductService)
		expectedStatus int
	}{
		{
			name: "Success",
			id:   "1",
			mockSetup: func(m *mockProductService) {
				m.getProductByIdFunc = func(ctx context.Context, id int) (db.Product, error) {
					if id != 1 {
						return db.Product{}, errors.New("unexpected id")
					}
					return db.Product{ID: 1, Name: "Test Product"}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Not Found",
			id:   "999",
			mockSetup: func(m *mockProductService) {
				m.getProductByIdFunc = func(ctx context.Context, id int) (db.Product, error) {
					return db.Product{}, internal.NewNotFoundError(id, "Product")
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid ID",
			id:             "abc",
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "DB Failure",
			id:   "1",
			mockSetup: func(m *mockProductService) {
				m.getProductByIdFunc = func(ctx context.Context, id int) (db.Product, error) {
					return db.Product{}, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockProductService{}
			tt.mockSetup(mock)
			h := &handler{service: mock}

			req := httptest.NewRequest(http.MethodGet, "/api/products/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			rr := httptest.NewRecorder()

			h.GetProductById(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				if rr.Header().Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %v", rr.Header().Get("Content-Type"))
				}
			}
		})
	}
}

func TestUpdateProduct(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		requestBody    any
		mockSetup      func(m *mockProductService)
		expectedStatus int
	}{
		{
			name: "Success",
			id:   "1",
			requestBody: updateProductRequest{
				PlatformID:  2,
				Name:        "Updated Product",
				Description: "Updated Description",
			},
			mockSetup: func(m *mockProductService) {
				m.updateProductFunc = func(ctx context.Context, req updateProductRequest, id int) (int, error) {
					if id != 1 {
						return 0, errors.New("unexpected id")
					}
					if req.PlatformID != 2 {
						return 0, errors.New("unexpected platform id")
					}
					if req.Name != "Updated Product" {
						return 0, errors.New("unexpected name")
					}
					return 1, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid ID",
			id:             "abc",
			requestBody:    updateProductRequest{PlatformID: 1, Name: "Test"},
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			id:             "1",
			requestBody:    "invalid json",
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing Name",
			id:   "1",
			requestBody: updateProductRequest{
				PlatformID: 1,
			},
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing PlatformID",
			id:   "1",
			requestBody: updateProductRequest{
				Name: "Updated Product",
			},
			mockSetup:      func(m *mockProductService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "DB Failure",
			id:   "1",
			requestBody: updateProductRequest{
				PlatformID: 2,
				Name:       "Fail Product",
			},
			mockSetup: func(m *mockProductService) {
				m.updateProductFunc = func(ctx context.Context, req updateProductRequest, id int) (int, error) {
					return 0, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Not Found",
			id:   "999",
			requestBody: updateProductRequest{
				PlatformID: 2,
				Name:       "Non-existent",
			},
			mockSetup: func(m *mockProductService) {
				m.updateProductFunc = func(ctx context.Context, req updateProductRequest, id int) (int, error) {
					return 0, internal.NewNotFoundError(id, "Product")
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockProductService{}
			tt.mockSetup(mock)
			h := &handler{service: mock}

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

			req := httptest.NewRequest(http.MethodPut, "/api/products/"+tt.id, bytes.NewBuffer(body))
			req.SetPathValue("id", tt.id)
			rr := httptest.NewRecorder()

			h.UpdateProduct(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}
		})
	}
}
