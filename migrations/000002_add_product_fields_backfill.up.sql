-- Add new columns to product table (no constraints, AutoMigrate will handle them)
ALTER TABLE `product`
ADD COLUMN `slug` VARCHAR(255) AFTER `id`,
ADD COLUMN `isActive` BOOLEAN NOT NULL DEFAULT TRUE AFTER `sku`,
ADD COLUMN `sortOrder` INT NOT NULL DEFAULT 0 AFTER `isActive`;

-- Backfill slugs from names for existing products
-- Handle empty/invariant slugs first
UPDATE `product` SET `slug` = LOWER(REPLACE(TRIM(`name`), ' ', '-')) WHERE `slug` = '' OR `slug` IS NULL;

-- Handle name conflicts by appending the ID
UPDATE `product` p1
JOIN (
    SELECT `id`, `slug`, ROW_NUMBER() OVER (PARTITION BY `slug` ORDER BY `id`) as `rn`
    FROM `product`
) p2 ON p1.`id` = p2.`id`
SET p1.`slug` = CONCAT(p1.`slug`, '-', p1.`id`)
WHERE p2.rn > 1;

-- Create unique index on slug after backfill
-- This is idempotent: if index already exists, this fails gracefully
ALTER TABLE `product`
ADD UNIQUE INDEX `idx_product_slug` (`slug`);