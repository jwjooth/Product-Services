package payload_test

import (
	"encoding/json"
	"testing"

	"golang-restful-api/entities"
	"golang-restful-api/payload"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductResponsePriceMarshalsAsString(t *testing.T) {
	resp := payload.ProductResponse{
		ID:    1,
		Name:  "Widget",
		Price: payload.Price{Cents: 1999},
	}
	b, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(b), `"price":"19.99"`)
}

func TestCreateProductRequestUnmarshalsPriceString(t *testing.T) {
	var req payload.CreateProductRequest
	err := json.Unmarshal([]byte(`{"name":"Widget","price":"19.99","stock":10,"sku":"ABC123"}`), &req)
	require.NoError(t, err)
	assert.Equal(t, int64(1999), req.Price.Cents)
}

func TestCreateProductRequestRejectsJSONNumberPrice(t *testing.T) {
	var req payload.CreateProductRequest
	err := json.Unmarshal([]byte(`{"name":"Widget","price":19.99,"stock":10,"sku":"ABC123"}`), &req)
	require.Error(t, err)
}

func TestUpdateProductRequestUnmarshalsPriceString(t *testing.T) {
	var req payload.UpdateProductRequest
	err := json.Unmarshal([]byte(`{"name":"Widget","price":"19.99","stock":10,"sku":"ABC123"}`), &req)
	require.NoError(t, err)
	assert.Equal(t, int64(1999), req.Price.Cents)
}

func TestProductFilterMinMaxAreInt64Pointers(t *testing.T) {
	min := int64(100)
	max := int64(20000)
	f := payload.ProductFilter{MinPrice: &min, MaxPrice: &max}
	assert.Equal(t, int64(100), *f.MinPrice)
	assert.Equal(t, int64(20000), *f.MaxPrice)
}

func TestNewProductResponseMapsPriceCents(t *testing.T) {
	e := entities.ProductEntity{
		ID:          1,
		Name:        "Widget",
		Description: "A widget",
		PriceCents:  2499,
		Category:    "Gadgets",
		ImageUrl:    "https://example.com/widget.jpg",
		Stock:       10,
		Rating:      4.5,
		ReviewCount: 3,
		SKU:         "WIDGET-001",
	}
	resp := payload.NewProductResponse(e)
	assert.Equal(t, int64(2499), resp.Price.Cents)
}
