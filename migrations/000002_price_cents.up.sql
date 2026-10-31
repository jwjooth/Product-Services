UPDATE `product` SET `price` = ROUND(`price` * 100);
ALTER TABLE `product` MODIFY COLUMN `price` BIGINT NOT NULL;
