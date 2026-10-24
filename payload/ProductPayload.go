package payload

import (
	"time"

	"golang-restful-api/entities"
)

// ProductResponse is the outward DTO for a product.
type ProductResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	ImageUrl    string    `json:"imageUrl"`
	Stock       int       `json:"stock"`
	Rating      float64   `json:"rating"`
	ReviewCount int       `json:"reviewCount"`
	SKU         string    `json:"sku"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProductRequest is the inbound DTO for POST /products.
type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description string  `json:"description,omitempty" validate:"max=5000"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"gte=0"`
	Category    string  `json:"category,omitempty" validate:"max=50"`
	ImageUrl    string  `json:"imageUrl,omitempty"`
	SKU         string  `json:"sku" validate:"required,max=50"`
}

// UpdateProductRequest is the inbound DTO for PUT /products/{id} (full replace).
type UpdateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description string  `json:"description,omitempty" validate:"max=5000"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"gte=0"`
	Category    string  `json:"category,omitempty" validate:"max=50"`
	ImageUrl    string  `json:"imageUrl,omitempty"`
	SKU         string  `json:"sku" validate:"required,max=50"`
}

// ListProductsMeta carries pagination metadata.
type ListProductsMeta struct {
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	Total      int64          `json:"total"`
	TotalPages int            `json:"total_pages"`
	Filters    *ProductFilter `json:"filters,omitempty"`
}

// ListProductsResponse is the paginated envelope for GET /products.
type ListProductsResponse struct {
	Data []ProductResponse `json:"data"`
	Meta ListProductsMeta  `json:"meta"`
}

// NewProductResponse maps an entity to its response DTO.
func NewProductResponse(e entities.ProductEntity) ProductResponse {
	return ProductResponse{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		Price:       e.Price,
		Category:    e.Category,
		ImageUrl:    e.ImageUrl,
		Stock:       e.Stock,
		Rating:      e.Rating,
		ReviewCount: e.ReviewCount,
		SKU:         e.SKU,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// NewProductResponses maps entities to response DTOs (never returns nil).
func NewProductResponses(list []entities.ProductEntity) []ProductResponse {
	out := make([]ProductResponse, 0, len(list))
	for _, e := range list {
		out = append(out, NewProductResponse(e))
	}
	return out
}

type ProductFilter struct {
	Category string
	MinPrice *float64
	MaxPrice *float64
	Search   string
	Sort     string
}

func SortClause(sort string) string {
	switch sort {
	case "price_asc":
		return "price ASC, id ASC"
	case "price_desc":
		return "price DESC, id ASC"
	case "name_asc":
		return "name ASC, id ASC"
	case "newest":
		return "createdAt DESC, id ASC"
	default:
		return "id ASC"
	}
}
