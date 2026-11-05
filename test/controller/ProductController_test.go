package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang-restful-api/controller"
	"golang-restful-api/helper"
	"golang-restful-api/payload"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) List(ctx context.Context, page, perPage int, filter payload.ProductFilter) ([]payload.ProductResponse, int64, *helper.AppError) {
	args := m.Called(ctx, page, perPage, filter)
	if args.Get(0) == nil {
		var err *helper.AppError
		if args.Get(2) != nil {
			err = args.Get(2).(*helper.AppError)
		}
		return nil, args.Get(1).(int64), err
	}
	var err *helper.AppError
	if args.Get(2) != nil {
		err = args.Get(2).(*helper.AppError)
	}
	return args.Get(0).([]payload.ProductResponse), args.Get(1).(int64), err
}

func (m *MockProductService) GetByID(ctx context.Context, id uint) (payload.ProductResponse, *helper.AppError) {
	args := m.Called(ctx, id)
	var err *helper.AppError
	if args.Get(1) != nil {
		err = args.Get(1).(*helper.AppError)
	}
	return args.Get(0).(payload.ProductResponse), err
}

func (m *MockProductService) Create(ctx context.Context, req payload.CreateProductRequest) (payload.ProductResponse, *helper.AppError) {
	args := m.Called(ctx, req)
	var err *helper.AppError
	if args.Get(1) != nil {
		err = args.Get(1).(*helper.AppError)
	}
	return args.Get(0).(payload.ProductResponse), err
}

func (m *MockProductService) Update(ctx context.Context, id uint, req payload.UpdateProductRequest) (payload.ProductResponse, *helper.AppError) {
	args := m.Called(ctx, id, req)
	var err *helper.AppError
	if args.Get(1) != nil {
		err = args.Get(1).(*helper.AppError)
	}
	return args.Get(0).(payload.ProductResponse), err
}

func (m *MockProductService) Delete(ctx context.Context, id uint) *helper.AppError {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*helper.AppError)
	}
	return nil
}

func setupProductRouter(ctrl controller.ProductController) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/products", ctrl.ListProducts)
	r.Post("/products", ctrl.CreateProduct)
	r.Get("/products/{id}", ctrl.GetProductByID)
	r.Put("/products/{id}", ctrl.UpdateProduct)
	r.Delete("/products/{id}", ctrl.DeleteProduct)
	return r
}

// TestProductController_ListProducts verifies product list responses and errors.
func TestProductController_ListProducts(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		expectedList := []payload.ProductResponse{
			{ID: 1, Name: "Keyboard", Price: payload.Price{Cents: 10000}, Stock: 5, Category: "Accessories", ImageUrl: "img.jpg", SKU: "KB001"},
		}
		svc.On("List", mock.Anything, 1, 10, payload.ProductFilter{}).Return(expectedList, int64(1), nil)

		req := httptest.NewRequest(http.MethodGet, "/products?page=1&per_page=10", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		svc.AssertExpectations(t)
	})

	t.Run("Service error", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		svc.On("List", mock.Anything, 1, 10, payload.ProductFilter{}).Return(nil, int64(0), helper.Internal("db failure", nil))

		req := httptest.NewRequest(http.MethodGet, "/products?page=1&per_page=10", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		svc.AssertExpectations(t)
	})

	t.Run("min_price exceeds max_price", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		req := httptest.NewRequest(http.MethodGet, "/products?min_price=100&max_price=50", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "min_price cannot exceed max_price")
		svc.AssertExpectations(t)
	})
}

// TestProductController_GetProductByID verifies product lookup responses and errors.
func TestProductController_GetProductByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		svc.On("GetByID", mock.Anything, uint(1)).Return(payload.ProductResponse{ID: 1, Name: "Mouse", Category: "Accessories", ImageUrl: "mouse.jpg", SKU: "MOUSE001"}, nil)

		req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		svc.AssertExpectations(t)
	})

	t.Run("Invalid ID param", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		req := httptest.NewRequest(http.MethodGet, "/products/abc", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Not found", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		svc.On("GetByID", mock.Anything, uint(99)).Return(payload.ProductResponse{}, helper.NotFound("product not found"))

		req := httptest.NewRequest(http.MethodGet, "/products/99", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		svc.AssertExpectations(t)
	})
}

// TestProductController_CreateProduct verifies product creation responses and errors.
func TestProductController_CreateProduct(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		createReq := payload.CreateProductRequest{Name: "Monitor", Price: payload.Price{Cents: 30000}, Stock: 4, Category: "Electronics", ImageUrl: "https://example.com/monitor.jpg", SKU: "MON001"}
		svc.On("Create", mock.Anything, createReq).Return(payload.ProductResponse{ID: 2, Name: "Monitor", Category: "Electronics", SKU: "MON001"}, nil)

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		svc.AssertExpectations(t)
	})

	t.Run("Invalid JSON body", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader([]byte("invalid json")))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Unknown JSON field", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader([]byte(`{"name":"Monitor","price":"300.00","stock":4,"sku":"MON001","unknown_prop":"hack"}`)))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), `unknown field`)
		assert.Contains(t, rec.Body.String(), `unknown_prop`)
	})

	t.Run("Multiple JSON values in body", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader([]byte(`{"name":"Monitor","price":"300.00","stock":4,"sku":"MON001"}{"extra":1}`)))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "multiple JSON values")
	})

	t.Run("Service validation error", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		createReq := payload.CreateProductRequest{Name: "", Price: payload.Price{Cents: 0}}
		svc.On("Create", mock.Anything, createReq).Return(payload.ProductResponse{}, helper.BadRequest("Name is required"))

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		svc.AssertExpectations(t)
	})
}

// TestProductController_UpdateProduct verifies product update responses and errors.
func TestProductController_UpdateProduct(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		updateReq := payload.UpdateProductRequest{Name: "Updated Monitor", Price: payload.Price{Cents: 35000}, Stock: 5, Category: "Electronics", ImageUrl: "https://example.com/updated_monitor.jpg", SKU: "MON002"}
		svc.On("Update", mock.Anything, uint(1), updateReq).Return(payload.ProductResponse{ID: 1, Name: "Updated Monitor", Category: "Electronics", SKU: "MON002"}, nil)

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest(http.MethodPut, "/products/1", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		svc.AssertExpectations(t)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		req := httptest.NewRequest(http.MethodPut, "/products/invalid", bytes.NewReader([]byte("{}")))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Invalid body", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		req := httptest.NewRequest(http.MethodPut, "/products/1", bytes.NewReader([]byte("not json")))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Unknown JSON field", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		req := httptest.NewRequest(http.MethodPut, "/products/1", bytes.NewReader([]byte(`{"name":"Monitor","price":"300.00","stock":4,"sku":"MON001","unknown_prop":"hack"}`)))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), `unknown field`)
		assert.Contains(t, rec.Body.String(), `unknown_prop`)
	})
}

func TestProductController_DeleteProduct(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		svc.On("Delete", mock.Anything, uint(1)).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/products/1", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		svc.AssertExpectations(t)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		req := httptest.NewRequest(http.MethodDelete, "/products/abc", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Not found", func(t *testing.T) {
		svc := new(MockProductService)
		ctrl := controller.NewProductController(svc)
		router := setupProductRouter(ctrl)

		svc.On("Delete", mock.Anything, uint(99)).Return(helper.NotFound("product not found"))

		req := httptest.NewRequest(http.MethodDelete, "/products/99", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		svc.AssertExpectations(t)
	})
}
