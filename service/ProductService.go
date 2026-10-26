package service

import (
	"context"
	"errors"

	"golang-restful-api/entities"
	"golang-restful-api/helper"
	"golang-restful-api/payload"
	"golang-restful-api/repository"

	"github.com/go-playground/validator/v10"
)

// ProductService holds business logic + validation and maps entities to DTOs.
type ProductService interface {
	List(ctx context.Context, page, perPage int, filter payload.ProductFilter) ([]payload.ProductResponse, int64, *helper.AppError)
	GetByID(ctx context.Context, id uint) (payload.ProductResponse, *helper.AppError)
	Create(ctx context.Context, req payload.CreateProductRequest) (payload.ProductResponse, *helper.AppError)
	Update(ctx context.Context, id uint, req payload.UpdateProductRequest) (payload.ProductResponse, *helper.AppError)
	Delete(ctx context.Context, id uint) *helper.AppError
}

type productService struct {
	repo     repository.ProductRepository
	validate *validator.Validate
}

// NewProductServiceImpl constructs a ProductService. It no longer needs *gorm.DB:
// transactions can be added later via a Tx-aware repository if required.
func NewProductServiceImpl(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo, validate: validator.New()}
}

// NewProductServiceImplWithDB keeps backward compatibility with the old
// constructor signature (db was unused). Prefer NewProductServiceImpl.
func NewProductServiceImplWithDB(_ any, repo repository.ProductRepository) ProductService {
	return NewProductServiceImpl(repo)
}

func (s *productService) List(ctx context.Context, page, perPage int, filter payload.ProductFilter) ([]payload.ProductResponse, int64, *helper.AppError) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	records, total, err := s.repo.FindAll(ctx, perPage, offset, filter)
	if err != nil {
		return nil, 0, helper.Internal("failed to fetch products", err)
	}
	return payload.NewProductResponses(records), total, nil
}

func (s *productService) GetByID(ctx context.Context, id uint) (payload.ProductResponse, *helper.AppError) {
	if id == 0 {
		return payload.ProductResponse{}, helper.BadRequest("invalid product id")
	}
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return payload.ProductResponse{}, helper.NotFound("product not found")
		}
		return payload.ProductResponse{}, helper.Internal("failed to fetch product", err)
	}
	return payload.NewProductResponse(*product), nil
}

// Create validates and persists a new product.
func (s *productService) Create(ctx context.Context, req payload.CreateProductRequest) (payload.ProductResponse, *helper.AppError) {
	if err := s.validate.Struct(req); err != nil {
		return payload.ProductResponse{}, helper.BadRequest(helper.FormatValidationError(err))
	}

	product := &entities.ProductEntity{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		ImageUrl:    req.ImageUrl,
		SKU:         req.SKU,
	}
	created, err := s.repo.Create(ctx, product)
	if err != nil {
		return payload.ProductResponse{}, helper.Internal("failed to create product", err)
	}
	return payload.NewProductResponse(*created), nil
}

// Update validates and replaces an existing product.
func (s *productService) Update(ctx context.Context, id uint, req payload.UpdateProductRequest) (payload.ProductResponse, *helper.AppError) {
	if id == 0 {
		return payload.ProductResponse{}, helper.BadRequest("invalid product id")
	}
	if err := s.validate.Struct(req); err != nil {
		return payload.ProductResponse{}, helper.BadRequest(helper.FormatValidationError(err))
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return payload.ProductResponse{}, helper.NotFound("product not found")
		}
		return payload.ProductResponse{}, helper.Internal("failed to fetch product", err)
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.Price = req.Price
	existing.Stock = req.Stock
	existing.Category = req.Category
	existing.ImageUrl = req.ImageUrl
	existing.SKU = req.SKU

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return payload.ProductResponse{}, helper.Internal("failed to update product", err)
	}
	return payload.NewProductResponse(*updated), nil
}

func (s *productService) Delete(ctx context.Context, id uint) *helper.AppError {
	if id == 0 {
		return helper.BadRequest("invalid product id")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("product not found")
		}
		return helper.Internal("failed to delete product", err)
	}
	return nil
}
