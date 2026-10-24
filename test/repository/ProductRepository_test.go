package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"golang-restful-api/entities"
	"golang-restful-api/payload"
	"golang-restful-api/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mock, sqlDB
}

// TestProductRepository_Create verifies inserting a product into the product table.
func TestProductRepository_Create(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repository.NewProductRepositoryImpl(gormDB)

	product := &entities.ProductEntity{
		Name:        "Gaming Mouse",
		Description: "Wireless mouse",
		Price:       49.99,
		Category:    "Electronics",
		ImageUrl:    "https://example.com/mouse.jpg",
		Stock:       10,
		SKU:         "MOUSE001",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `product` (`name`,`description`,`category`,`price`,`imageUrl`,`stock`,`rating`,`reviewCount`,`sku`,`createdAt`,`updatedAt`) VALUES (?,?,?,?,?,?,?,?,?,?,?)")).
		WithArgs(product.Name, product.Description, product.Category, product.Price, product.ImageUrl, product.Stock, product.Rating, product.ReviewCount, product.SKU, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	result, err := repo.Create(context.Background(), product)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), result.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestProductRepository_FindAll verifies listing products and count failures.
func TestProductRepository_FindAll(t *testing.T) {
	t.Run("Success with records and count", func(t *testing.T) {
		gormDB, mock, sqlDB := setupMockDB(t)
		defer sqlDB.Close()

		repo := repository.NewProductRepositoryImpl(gormDB)

		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `product`")).
			WillReturnRows(countRows)

		productRows := sqlmock.NewRows([]string{"id", "name", "description", "price", "category", "imageUrl", "stock", "rating", "reviewCount", "sku", "createdAt", "updatedAt"}).
			AddRow(1, "Prod 1", "Desc 1", 10.0, "Electronics", "img1.jpg", 5, 4.5, 10, "SKU001", time.Now(), time.Now()).
			AddRow(2, "Prod 2", "Desc 2", 20.0, "Home", "img2.jpg", 8, 4.0, 5, "SKU002", time.Now(), time.Now())
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `product` ORDER BY id ASC LIMIT ?")).
			WithArgs(10).
			WillReturnRows(productRows)

		records, total, err := repo.FindAll(context.Background(), 10, 0, payload.ProductFilter{})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, records, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Filters and sort", func(t *testing.T) {
		gormDB, mock, sqlDB := setupMockDB(t)
		defer sqlDB.Close()

		repo := repository.NewProductRepositoryImpl(gormDB)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `product`")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `product` WHERE category = ? AND price <= ? AND price >= ? AND (name LIKE ? OR sku LIKE ?) ORDER BY price ASC, id ASC LIMIT ? OFFSET ?")).
			WithArgs("Electronics", 100.0, 10.0, "%mouse%", "%mouse%", 5, 10).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Gaming Mouse"))

		minPrice, maxPrice := 10.0, 100.0
		records, total, err := repo.FindAll(context.Background(), 5, 10, payload.ProductFilter{
			Category: "Electronics",
			MinPrice: &minPrice,
			MaxPrice: &maxPrice,
			Search:   "mouse",
			Sort:     "price_asc",
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, records, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Count query fails", func(t *testing.T) {
		gormDB, mock, sqlDB := setupMockDB(t)
		defer sqlDB.Close()

		repo := repository.NewProductRepositoryImpl(gormDB)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `product`")).
			WillReturnError(errors.New("db error"))

		records, total, err := repo.FindAll(context.Background(), 10, 0, payload.ProductFilter{})
		assert.Error(t, err)
		assert.Equal(t, int64(0), total)
		assert.Nil(t, records)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestProductRepository_FindByID verifies successful and missing product lookups.
func TestProductRepository_FindByID(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		gormDB, mock, sqlDB := setupMockDB(t)
		defer sqlDB.Close()

		repo := repository.NewProductRepositoryImpl(gormDB)

		row := sqlmock.NewRows([]string{"id", "name", "description", "price", "category", "imageUrl", "stock", "rating", "reviewCount", "sku", "created_at", "updated_at"}).
			AddRow(1, "Prod 1", "Desc 1", 10.0, "Electronics", "img.jpg", 5, 4.5, 10, "SKU001", time.Now(), time.Now())
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `product` WHERE `product`.`id` = ? ORDER BY `product`.`id` LIMIT ?")).
			WithArgs(1, 1).
			WillReturnRows(row)

		product, err := repo.FindByID(context.Background(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, uint(1), product.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Not found", func(t *testing.T) {
		gormDB, mock, sqlDB := setupMockDB(t)
		defer sqlDB.Close()

		repo := repository.NewProductRepositoryImpl(gormDB)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `product` WHERE `product`.`id` = ? ORDER BY `product`.`id` LIMIT ?")).
			WithArgs(99, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		product, err := repo.FindByID(context.Background(), 99)
		assert.ErrorIs(t, err, repository.ErrNotFound)
		assert.Nil(t, product)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestProductRepository_Update verifies updating a product in the product table.
func TestProductRepository_Update(t *testing.T) {
	gormDB, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := repository.NewProductRepositoryImpl(gormDB)

	product := &entities.ProductEntity{
		ID:          1,
		Name:        "Updated Mouse",
		Description: "Updated desc",
		Price:       59.99,
		Category:    "Electronics",
		ImageUrl:    "https://example.com/updated_mouse.jpg",
		Stock:       15,
		Rating:      4.5,
		ReviewCount: 12,
		SKU:         "MOUSE002",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `product` SET `name`=?,`description`=?,`category`=?,`price`=?,`imageUrl`=?,`stock`=?,`rating`=?,`reviewCount`=?,`sku`=?,`createdAt`=?,`updatedAt`=? WHERE `id` = ?")).
		WithArgs(product.Name, product.Description, product.Category, product.Price, product.ImageUrl, product.Stock, product.Rating, product.ReviewCount, product.SKU, sqlmock.AnyArg(), sqlmock.AnyArg(), product.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	updated, err := repo.Update(context.Background(), product)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "Updated Mouse", updated.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestProductRepository_Delete verifies successful and missing product deletion.
func TestProductRepository_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		gormDB, mock, sqlDB := setupMockDB(t)
		defer sqlDB.Close()

		repo := repository.NewProductRepositoryImpl(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `product` WHERE `product`.`id` = ?")).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), 1)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Not found (RowsAffected = 0)", func(t *testing.T) {
		gormDB, mock, sqlDB := setupMockDB(t)
		defer sqlDB.Close()

		repo := repository.NewProductRepositoryImpl(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `product` WHERE `product`.`id` = ?")).
			WithArgs(99).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), 99)
		assert.ErrorIs(t, err, repository.ErrNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
