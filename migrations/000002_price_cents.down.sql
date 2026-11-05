ALTER TABLE `product` MODIFY COLUMN `price` DOUBLE NOT NULL;
UPDATE `product` SET `price` = `price` / 100;
