CREATE TABLE IF NOT EXISTS `product` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL,
    `description` TEXT,
    `category` VARCHAR(50),
    `price` DOUBLE NOT NULL,
    `imageUrl` TEXT,
    `stock` BIGINT NOT NULL DEFAULT 0,
    `rating` DECIMAL(3,2),
    `reviewCount` BIGINT DEFAULT 0,
    `sku` VARCHAR(50),
    `createdAt` DATETIME(3),
    `updatedAt` DATETIME(3),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_product_sku` (`sku`),
    KEY `idx_category_price` (`category`, `price`)
);
