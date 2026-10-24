package controller

import (
	"encoding/json"
	"net/http"

	"golang-restful-api/helper"
	"golang-restful-api/payload"
	"golang-restful-api/service"
)

// ProductController is the HTTP boundary: decode → service → encode.
// No business logic or SQL here.
type ProductController interface {
	ListProducts(w http.ResponseWriter, r *http.Request)
	GetProductByID(w http.ResponseWriter, r *http.Request)
	CreateProduct(w http.ResponseWriter, r *http.Request)
	UpdateProduct(w http.ResponseWriter, r *http.Request)
	DeleteProduct(w http.ResponseWriter, r *http.Request)
}

type productController struct {
	service service.ProductService
}

// NewProductController constructs a ProductController.
func NewProductController(svc service.ProductService) ProductController {
	return &productController{service: svc}
}

// ListProducts @Summary List all products with pagination
// @Description Retrieve a paginated list of all products
// @Tags product
// @Accept json
// @Produce json
// @Param page query int false "Page number (1-based)" minimum(1) default(1)
// @Param per_page query int false "Items per page (max 100)" minimum(1) maximum(100) default(10)
// @Param category query string false "Filter by category"
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Param q query string false "Search query (name or SKU)"
// @Param sort query string false "Sort order: price_asc, price_desc, name_asc, newest"
// @Success 200 {object} helper.SuccessEnvelope{Data=[]payload.ProductResponse,Meta=payload.ListProductsMeta}
// @Failure 500 {object} helper.ErrorEnvelope
// @Router /api/v1/products [get]
func (c *productController) ListProducts(w http.ResponseWriter, r *http.Request) {
	page := helper.QueryInt(r, "page", 1)
	perPage := helper.QueryInt(r, "per_page", 10)

	items, total, appErr := c.service.List(r.Context(), page, perPage)
	if appErr != nil {
		helper.WriteAppError(w, appErr)
		return
	}

	totalPages := 0
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}
	meta := payload.ListProductsMeta{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}
	helper.WriteSuccessWithMeta(w, http.StatusOK, "success", items, meta)
}

// GetProductByID @Summary Get a product by ID
// @Description Retrieve a single product by its unique ID
// @Tags product
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Failure 400 {object} helper.ErrorEnvelope
// @Failure 404 {object} helper.ErrorEnvelope
// @Router /api/v1/products/{id} [get]
// @Success 200 {object} helper.SuccessEnvelope{Data=payload.ProductResponse}
func (c *productController) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id, appErr := helper.ParseIDParam(r, "id")
	if appErr != nil {
		helper.WriteAppError(w, appErr)
		return
	}
	result, err := c.service.GetByID(r.Context(), id)
	if err != nil {
		helper.WriteAppError(w, err)
		return
	}
	helper.WriteSuccess(w, http.StatusOK, "success", result)
}

// CreateProduct @Summary Create a new product
// @Description Create a new product with the provided details
// @Tags product
// @Accept json
// @Produce json
// @Param body body payload.CreateProductRequest true "Product request payload"
// @Failure 400 {object} helper.ErrorEnvelope
// @Failure 500 {object} helper.ErrorEnvelope
// @Router /api/v1/products [post]
// @Success 201 {object} helper.SuccessEnvelope{Data=payload.ProductResponse}
func (c *productController) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req payload.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	result, appErr := c.service.Create(r.Context(), req)
	if appErr != nil {
		helper.WriteAppError(w, appErr)
		return
	}
	helper.WriteSuccess(w, http.StatusCreated, "success", result)
}

// UpdateProduct @Summary Update an existing product
// @Description Update an existing product by its ID with the provided details
// @Tags product
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param body body payload.UpdateProductRequest true "Product update payload"
// @Failure 400 {object} helper.ErrorEnvelope
// @Failure 404 {object} helper.ErrorEnvelope
// @Failure 500 {object} helper.ErrorEnvelope
// @Router /api/v1/products/{id} [put]
// @Success 200 {object} helper.SuccessEnvelope{Data=payload.ProductResponse}
func (c *productController) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, appErr := helper.ParseIDParam(r, "id")
	if appErr != nil {
		helper.WriteAppError(w, appErr)
		return
	}
	var req payload.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	result, err := c.service.Update(r.Context(), id, req)
	if err != nil {
		helper.WriteAppError(w, err)
		return
	}
	helper.WriteSuccess(w, http.StatusOK, "success", result)
}

// DeleteProduct @Summary Delete a product by ID
// @Description Delete a product by its unique ID
// @Tags product
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Failure 400 {object} helper.ErrorEnvelope
// @Failure 404 {object} helper.ErrorEnvelope
// @Failure 500 {object} helper.ErrorEnvelope
// @Router /api/v1/products/{id} [delete]
// @Success 200 {object} helper.SuccessEnvelope
func (c *productController) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, appErr := helper.ParseIDParam(r, "id")
	if appErr != nil {
		helper.WriteAppError(w, appErr)
		return
	}
	if err := c.service.Delete(r.Context(), id); err != nil {
		helper.WriteAppError(w, err)
		return
	}
	helper.WriteSuccess(w, http.StatusOK, "product deleted successfully", nil)
}
