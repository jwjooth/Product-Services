# Price as integer cents (#35) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace `float64` price representation throughout `entities/`, `payload/`, and `service/` with an integer number of cents (`PriceCents int64`), transported as a JSON string (`"19.99"` → 1999 cents), with exact validation and cent-based price filtering.

**Architecture:** Store price as BIGINT cents in MySQL; entity field becomes `PriceCents int64`. Payload gains a custom `Price` type marshalled/unmarshalled as a JSON string with 2-decimal-place validation. Service performs all arithmetic/comparison in cents. Repository converts filter min/max prices to cents before SQL. Swagger docs regenerated. CHANGELOG entry added. No new third-party dependency (no `shopspring/decimal`), per issue recommendation.

**Tech Stack:** Go 1.26.5, standard library (`math`/`math/big` not needed for two-decimal parsing; custom int64 parsing is enough), existing test deps (testify, sqlmock).

**Spec:** https://github.com/jwjooth/Product-Services/issues/35 (issue body is the full design + acceptance criteria + test plan)

## Global Constraints

- No price represented as `float64` in `entities/`, `payload/`, or `service/` (verbatim from acceptance criteria).
- POST with `"price": "19.99"` → persists 1999 cents.
- GET returns `"price": "19.99"` as a JSON string.
- POST with `price: 19.99` JSON number → rejected with clear 400.
- `"price": "19.999"` (3 decimal places) → rejected with clear 400.
- Negative prices → rejected.
- Absurdly large price → rejected (overflow guard).
- Stored column type is `BIGINT` (cents).
- Existing rows migrated correctly (`SHOW CREATE TABLE product`, spot-check `19.99` → `1999` cents).
- Filter min/max price compares in cents, no float comparison.
- Swagger spec documents price format and currency.
- Migration script included for existing data.

## Review Focus

- JSON string decoding of `price` with custom `json.Unmarshaler` accepting only strings (rejects JSON number input).
- Parsing `"19.99"` → 1999 cents exactly; rejection of `"19.999"` (more than 2 decimals).
- Rejection of negative values and overflow-causing values (> 1,000,000,0000 cents = $100M).
- Repository price filter converts cents correctly; cent comparison in SQL.
- Migration correctness for existing `double` column → `BIGINT` cents column with `ROUND(price * 100)` data copy.

---

## Task 1: Write `payload/Price.go` — custom monetary type

**Files:**
- Create: `payload/Price.go`
- Test: `test/payload/Price_test.go`

**Interfaces:**
- Produces: `type Price struct` with `Cents int64`, `Currency string`, `MarshalJSON`, `UnmarshalJSON`, `NewPriceFromCents(int64) (Price, error)`, `ParsePrice(string) (Price, error)`, `String() string`.

- [ ] **Step 1: Write failing tests**

```go
// test/payload/Price_test.go
func TestPriceMarshalUnmarshalRoundTrip(t *testing.T) {
    p, err := NewPriceFromCents(1999)
    assert.NoError(t, err)
    assert.Equal(t, "19.99", p.String())
    b, err := json.Marshal(p)
    assert.NoError(t, err)
    assert.Equal(t, `"19.99"`, string(b))
    var out Price
    assert.NoError(t, json.Unmarshal(b, &out))
    assert.Equal(t, int64(1999), out.Cents)
}

func TestPriceRejectsJSONNumber(t *testing.T) {
    var p Price
    err := json.Unmarshal([]byte(`{"price":19.99}`), &p)
    assert.Error(t, err)
}

func TestPriceRejectsThreeDecimals(t *testing.T) {
    _, err := ParsePrice("19.999")
    assert.Error(t, err)
}

func TestPriceRejectsNegative(t *testing.T) {
    _, err := NewPriceFromCents(-100)
    assert.Error(t, err)
}

func TestPriceRejectsOverflow(t *testing.T) {
    _, err := NewPriceFromCents(10000000000) // $100M ceiling
    assert.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails** — `go test ./test/payload/...` → FAIL (file doesn't exist).

- [ ] **Step 3: Implement `payload/Price.go`**

Custom `MarshalJSON`/`UnmarshalJSON` so JSON numbers are rejected; `ParsePrice` parses string, validates ≤ 2 decimals, non-negative, under ceiling ($100M). Use `strings`/`strconv` only — no third-party dep.

- [ ] **Step 4: Run test to verify it passes** — `go test ./test/payload/...` → PASS.

- [ ] **Step 5: Commit** — `git add payload/Price.go test/payload/Price_test.go` → `git commit -m "feat(payload): add Price monetary type in cents"`.

---

## Task 2: Replace `Price float64` with `PriceCents int64` in entity

**Files:**
- Modify: `entities/ProductEntity.go` line 14: `Price float64 ...` → `PriceCents int64` with `gorm:"column:price;not null"` (keep column name `price`, change type to BIGINT via migration).

**Interfaces:**
- Consumes: nothing new.
- Produces: `ProductEntity.PriceCents int64` field used by service/repository/tests.

- [ ] **Step 1: Write failing test** — existing `test/repository/ProductRepository_test.go` references `Price` field; update after entity change. No new test file needed; task verified via compilation.

- [ ] **Step 2: Compile-check** — `go build ./...` → FAIL until dependent tasks fix references.

- [ ] **Step 3: Modify entity** — one-line field change plus optional gorm column comment. Keep `price` column name so migration can ALTER type + copy data.

- [ ] **Step 4: Run compilation** — after Task 4/5 fix references, `go build ./...` → PASS.

- [ ] **Step 5: Commit** — `git add entities/ProductEntity.go` → `git commit -m "refactor(entity): replace Price float64 with PriceCents int64"`.

---

## Task 3: Replace `Price float64` fields in payload requests/responses

**Files:**
- Modify: `payload/ProductPayload.go` — `ProductResponse.Price`, `CreateProductRequest.Price`, `UpdateProductRequest.Price` → `Price Price` (custom type).
- Modify: `payload/ProductFilter` — `MinPrice`/`MaxPrice` → `*int64` (cents).

**Interfaces:**
- Produces: `ProductResponse.Price` marshals as `"19.99"`; `CreateProductRequest.Price` unmarshals string → rejects number; `ProductFilter.MinPrice`/`MaxPrice` are `*int64` cents.

- [ ] **Step 1: Write failing test** — compile check via `go build ./...` → FAIL until field types change.

- [ ] **Step 2: Modify payload types** — swap field types, add `"price"` string tag with `json:"price"`.

- [ ] **Step 3: Compile-check** — `go build ./...` → FAIL (service/repository/controller reference old fields).

- [ ] **Step 4: Commit** — `git add payload/ProductPayload.go` → `git commit -m "refactor(payload): Price fields use Price type in cents"`.

---

## Task 4: Update repository price filter to cents comparison

**Files:**
- Modify: `repository/ProductRepository.go` `applyProductFilters` — convert `*int64` cents directly into SQL `price <= ?` / `price >= ?`.

**Interfaces:**
- Consumes: `payload.ProductFilter` with `*int64` cents.
- Produces: SQL comparison `price <= ?` / `price >= ?` with cent integers.

- [ ] **Step 1: Modify filter application** — pass `*filter.MinPrice` / `*filter.MaxPrice` directly as int64 args.

- [ ] **Step 2: Update existing repository test expectations** — `TestProductRepository_FindAll` mock expectations change `"SELECT count(*) FROM product WHERE category = ? AND price <= ? ..."` args to cents ints.

- [ ] **Step 3: Run tests** — `go test ./test/repository/...` → PASS.

- [ ] **Step 4: Commit** — `git add repository/ProductRepository.go test/repository/ProductRepository_test.go` → `git commit -m "fix(repository): price filter compares in cents"`.

---

## Task 5: Update service price mapping

**Files:**
- Modify: `service/ProductService.go` — `Create`/`Update` map `req.Price.Cents` → `product.PriceCents`; nothing else changes (entity handles storage).

**Interfaces:**
- Consumes: `payload.CreateProductRequest.Price` of type `Price`.
- Produces: `entities.ProductEntity.PriceCents int64`.

- [ ] **Step 1: Modify service mapping** — `Price: product.PriceCents` etc.

- [ ] **Step 2: Compile-check + tests** — `go test ./test/service/...` → PASS after service test updates in Task 6.

- [ ] **Step 3: Commit** — `git add service/ProductService.go` → `git commit -m "refactor(service): map Price type to entity PriceCents"`.

---

## Task 6: Update controller min/max price parsing

**Files:**
- Modify: `controller/ProductController.go` — `QueryFloat` → `QueryInt` for min/max price; filter `MinPrice`/`MaxPrice` become `*int64` cents.

**Interfaces:**
- Consumes: query params `min_price`/`max_price` as integer cents.
- Produces: `payload.ProductFilter` with `*int64` cents.

- [ ] **Step 1: Modify controller** — parse `min_price`/`max_price` with `QueryInt`, store as `*int64`.

- [ ] **Step 2: Compile-check** — `go build ./...` → PASS.

- [ ] **Step 3: Commit** — `git add controller/ProductController.go` → `git commit -m "refactor(controller): min/max price filters use cents"`.

---

## Task 7: Add migration script + update tests across all test packages

**Files:**
- Create: `migrations/000002_price_cents.up.sql` and `.down.sql`.
- Modify: `test/repository/ProductRepository_test.go` — mock SQL strings/args updated to cents (price column now BIGINT int64).
- Modify: `test/service/ProductService_test.go` — request Price as `"19.99"` string, assertions on cents.
- Modify: `test/controller/ProductController_test.go` — request body price as JSON string `"19.99"`.
- Modify: `test/payload/Price_test.go` (created in Task 1).

**Interfaces:**
- Migration: `ALTER TABLE product MODIFY price BIGINT NOT NULL; UPDATE product SET price = ROUND(price * 100);` (down reverses).
- Test payload: `CreateProductRequest{Price: ParsePrice("19.99")}` → `PriceCents == 1999`.

- [ ] **Step 1: Write migration SQL** — up: ALTER COLUMN type + data copy; down: revert.

- [ ] **Step 2: Update repository test mock expectations** — price filter args become int64 cents.

- [ ] **Step 3: Update service/controller tests** — price input as `"19.99"` string; assert `Price` field serializes as `"19.99"`.

- [ ] **Step 4: Run full suite** — `go test ./...` → PASS.

- [ ] **Step 5: Commit** — `git add migrations/ test/...` → `git commit -m "test(repo,service,controller): price cents round-trip + migration script"`.

---

## Task 8: Regenerate Swagger docs + update main.go annotation

**Files:**
- Modify: `main.go` line 28 `@version 1.0.0` → `@version 1.2.0` (per milestone v1.2).
- Regenerate: `swag init -g main.go -o docs` (CI verifies docs are current).
- Modify: `docs/swagger.json`, `docs/swagger.yaml`, `docs/docs.go` — price field becomes string with format/description about cents.

**Interfaces:**
- Swagger `price` parameter/response field type: string.

- [ ] **Step 1: Update swagger annotation** — `@version 1.2.0`; add `@Param` description noting price is a string `"19.99"`.

- [ ] **Step 2: Regenerate docs** — `swag init -g main.go -o docs`.

- [ ] **Step 3: Verify CI check passes** — `git diff --quiet -- docs/` → clean (docs committed).

- [ ] **Step 4: Commit** — `git add main.go docs/` → `git commit -m "docs(swagger): v1.2.0 price as string cents"`.

---

## Task 9: Add CHANGELOG entry + AGENTS.md note

**Files:**
- Modify: `CHANGELOG.md` — add `1.2.0` section with price refactor.
- Modify: `AGENTS.md` — note price type change (optional, keeps agent guidance accurate).

**Interfaces:**
- None.

- [ ] **Step 1: Edit CHANGELOG.md** — new unreleased `1.2.0` section listing breaking API change (price number → string).

- [ ] **Step 2: Commit** — `git add CHANGELOG.md AGENTS.md` → `git commit -m "docs(release): v1.2.0 price cents changelog entry"`.

---

## Self-review

1. **Spec coverage:**
   - No float64 price in entities/payload/service → Tasks 1-5.
   - POST `"19.99"` → 1999 cents → Task 1 + 3.
   - GET returns `"19.99"` string → Task 1 MarshalJSON.
   - POST number `19.99` → rejected 400 → Task 1 UnmarshalJSON rejects JSON numbers; controller decodes strict.
   - `"19.999"` 3 decimals → rejected → Task 1 ParsePrice.
   - Negative → rejected → Task 1.
   - Absurdly large → rejected → Task 1 ceiling check ($100M).
   - Stored column BIGINT → migration Task 7.
   - Existing rows migrated `ROUND(price*100)` → migration Task 7.
   - Filter in cents → Task 4 + 6.
   - Swagger documents price format → Task 8.
   - Migration script → Task 7.

2. **Step scan:** Every step lets the implementer write exactly one reasonable thing; no stubs like "handle edge cases".

3. **Type consistency:** `PriceCents int64` entity ↔ `Price{Cents}` payload ↔ `*int64` filter, all passing through service/controller by Task 5/6.

4. **Review Focus:** each failure mode maps to a task with tests: JSON-number rejection (Task 1), 3-decimal rejection (Task 1), overflow (Task 1), negative (Task 1), cent filter comparison (Task 4), migration data copy (Task 7).

5. **Proportion:** Plan ~match spec complexity (single focused refactor), no transcript bodies.

---

Plan complete and saved to `docs/superpowers/plans/2026-10-31-fix-price-float64-cents.md`. Please review the plan. Does it capture what you want?
