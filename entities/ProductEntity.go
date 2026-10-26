package entities

import "time"

// TableNameProduct is the database table used for product records.
const TableNameProduct = "product"

// ProductEntity is the persistence model for the product table.
type ProductEntity struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string    `gorm:"column:name;size:255;not null" json:"name"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	Category    string    `gorm:"column:category;size:50;index:idx_category_price"`
	Price       float64   `gorm:"column:price;not null;index:idx_category_price"`
	ImageUrl    string    `gorm:"column:imageUrl;type:text" json:"imageUrl"`
	Stock       int       `gorm:"column:stock;not null;default:0" json:"stock"`
	Rating      float64   `gorm:"column:rating;type:decimal(3,2)" json:"rating"`
	ReviewCount int       `gorm:"column:reviewCount;default:0" json:"reviewCount"`
	SKU         string    `gorm:"column:sku;unique;size:50" json:"sku"`
	CreatedAt   time.Time `gorm:"column:createdAt;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updatedAt;autoUpdateTime" json:"updated_at"`
}

// TableName specifies the product table name for GORM.
func (*ProductEntity) TableName() string {
	return TableNameProduct
}
