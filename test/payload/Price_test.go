package payload_test

import (
	"encoding/json"
	"testing"

	"golang-restful-api/payload"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPriceMarshalUnmarshalRoundTrip(t *testing.T) {
	p, err := payload.NewPriceFromCents(1999)
	require.NoError(t, err)
	assert.Equal(t, "19.99", p.String())
	assert.Equal(t, int64(1999), p.Cents)

	b, err := json.Marshal(p)
	require.NoError(t, err)
	assert.Equal(t, `"19.99"`, string(b))

	var out payload.Price
	require.NoError(t, json.Unmarshal(b, &out))
	assert.Equal(t, int64(1999), out.Cents)
}

func TestPriceCentsIsJSONInvisible(t *testing.T) {
	type wrapper struct {
		Price payload.Price `json:"price"`
	}
	w, err := payload.NewPriceFromCents(1234)
	require.NoError(t, err)

	b, err := json.Marshal(wrapper{Price: w})
	require.NoError(t, err)
	assert.Equal(t, `{"price":"12.34"}`, string(b))
	assert.NotContains(t, string(b), "1234")
}

func TestPriceRejectsJSONNumber(t *testing.T) {
	type wrapper struct {
		Price payload.Price `json:"price"`
	}

	// A JSON number inside a Price field must be rejected (real wrapper-object scenario).
	var w wrapper
	err := json.Unmarshal([]byte(`{"price":19.99}`), &w)
	require.Error(t, err)

	// A bare JSON number must be rejected when unmarshaled into Price directly.
	var p payload.Price
	err = json.Unmarshal([]byte(`19.99`), &p)
	require.Error(t, err)
}

func TestParsePrice(t *testing.T) {
	cases := map[string]int64{
		"19.99":  1999,
		"19":     1900,
		"19.9":   1990,
		"19.00":  1900,
		"0.01":   1,
		"0":      0,
		"100.00": 10000,
	}
	for in, want := range cases {
		p, err := payload.ParsePrice(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, p.Cents, in)
	}
}

func TestPriceRejectsThreeDecimals(t *testing.T) {
	_, err := payload.ParsePrice("19.999")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "19.999")
}

func TestPriceRejectsNegative(t *testing.T) {
	_, err := payload.NewPriceFromCents(-100)
	require.Error(t, err)

	_, err = payload.ParsePrice("-19.99")
	require.Error(t, err)
}

func TestPriceRejectsOverflow(t *testing.T) {
	_, err := payload.NewPriceFromCents(10000000000) // $100M ceiling
	require.Error(t, err)

	_, err = payload.ParsePrice("100000000.00")
	require.Error(t, err)

	// Just below the ceiling is accepted.
	p, err := payload.NewPriceFromCents(9999999999)
	require.NoError(t, err)
	assert.Equal(t, "99999999.99", p.String())
}

func TestPriceRejectsGarbage(t *testing.T) {
	for _, in := range []string{"", "abc", "1.2.3", "19,99", " 19.99", "1e3", "+19.99"} {
		_, err := payload.ParsePrice(in)
		assert.Error(t, err, in)
	}
}
