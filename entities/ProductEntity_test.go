package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductEntity_PriceCentsField(t *testing.T) {
	entity := ProductEntity{
		PriceCents: 1999,
	}
	assert.Equal(t, int64(1999), entity.PriceCents)
}
