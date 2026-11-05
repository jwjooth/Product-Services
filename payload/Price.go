package payload

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Price represents a monetary amount transported as a JSON string (e.g. "19.99").
// @Schema type=string format=string
type Price struct {
	Cents int64 `json:"-"`
}

const maxPriceCents = 10_000_000_000 // $100M, exclusive ceiling

func NewPriceFromCents(cents int64) (Price, error) {
	if cents < 0 {
		return Price{}, fmt.Errorf("invalid price: %d cents is negative", cents)
	}
	if cents >= maxPriceCents {
		return Price{}, fmt.Errorf("invalid price: %d cents exceeds maximum of %d", cents, maxPriceCents)
	}
	return Price{Cents: cents}, nil
}

func ParsePrice(s string) (Price, error) {
	// Reject leading/trailing whitespace.
	if s != strings.TrimSpace(s) {
		return Price{}, fmt.Errorf("invalid price %q: leading or trailing whitespace not allowed", s)
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return Price{}, fmt.Errorf("invalid price %q: empty", s)
	}
	// Reject negative sign early for clearer error.
	if strings.HasPrefix(s, "-") {
		return Price{}, fmt.Errorf("invalid price %q: negative prices not allowed", s)
	}
	// Split integer and fractional parts.
	var whole, frac string
	parts := strings.SplitN(s, ".", 2)
	if len(parts) == 1 {
		whole = parts[0]
		frac = "00"
	} else {
		whole, frac = parts[0], parts[1]
	}
	if whole == "" {
		return Price{}, fmt.Errorf("invalid price %q: missing integer part", s)
	}
	// Validate each part consists of decimal digits.
	if !onlyDigits(whole) {
		return Price{}, fmt.Errorf("invalid price %q: integer part must contain only digits", s)
	}
	if len(frac) > 2 {
		return Price{}, fmt.Errorf("invalid price %q: at most 2 decimal places allowed", s)
	}
	if !onlyDigits(frac) {
		return Price{}, fmt.Errorf("invalid price %q: fractional part must contain only digits", s)
	}
	// Pad fractional part to two digits.
	if len(frac) == 1 {
		frac += "0"
	} else if len(frac) == 0 {
		frac = "00"
	}
	// Parse whole part safely (up to maxPriceCents/100).
	maxWhole := maxPriceCents / 100
	if maxWhole > 0 && len(whole) > int(maxWhole) { // rough length check to avoid overflow later
		return Price{}, fmt.Errorf("invalid price %q: exceeds maximum amount", s)
	}
	wholeVal, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return Price{}, fmt.Errorf("invalid price %q: invalid integer part", s)
	}
	fracVal, _ := strconv.ParseInt(frac, 10, 64)
	// Final overflow guard: whole*100 + frac may exceed maxPriceCents due to whole close to ceiling.
	if wholeVal > maxPriceCents/100 {
		return Price{}, fmt.Errorf("invalid price %q: exceeds maximum amount", s)
	}
	if wholeVal*100+fracVal >= maxPriceCents {
		return Price{}, fmt.Errorf("invalid price %q: exceeds maximum amount", s)
	}
	return NewPriceFromCents(wholeVal*100 + fracVal)
}

func onlyDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (p Price) String() string {
	// Negative cents are prohibited by NewPriceFromCents, so this is safe.
	return fmt.Sprintf("%d.%02d", p.Cents/100, p.Cents%100)
}

func (p Price) MarshalJSON() ([]byte, error) {
	// The JSON tag for Cents is "-", but custom MarshalJSON overrides field handling.
	return json.Marshal(p.String())
}

func (p *Price) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("invalid price %q: empty", data)
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("invalid price: expected a JSON string like \"19.99\", got %s", string(data))
	}
	if s == "" {
		return fmt.Errorf("invalid price: empty")
	}
	parsed, err := ParsePrice(s)
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}
