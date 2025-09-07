-- Basic stored procedure
CREATE PROCEDURE GetAllUsers() BEGIN SELECT * FROM users; END;

-- Procedure with parameters
CREATE PROCEDURE UpdateUserStatus(IN user_id INT, IN new_status VARCHAR(20)) BEGIN UPDATE users SET status = new_status WHERE id = user_id; END;

-- Procedure with IF/ELSE control flow
CREATE PROCEDURE CheckInventory(IN product_id INT, OUT status VARCHAR(20)) BEGIN DECLARE stock INT; SELECT quantity INTO stock FROM inventory WHERE id = product_id; IF stock > 100 THEN SET status = 'In Stock'; ELSEIF stock > 0 THEN SET status = 'Low Stock'; ELSE SET status = 'Out of Stock'; END IF; END;

-- Procedure with WHILE loop
CREATE PROCEDURE GenerateSequence(IN max_value INT) BEGIN DECLARE counter INT DEFAULT 1; CREATE TEMPORARY TABLE sequence_table (value INT); WHILE counter <= max_value DO INSERT INTO sequence_table VALUES (counter); SET counter = counter + 1; END WHILE; SELECT * FROM sequence_table; DROP TEMPORARY TABLE sequence_table; END;

-- Procedure with cursor and handler
CREATE PROCEDURE CalculateTotalSales() BEGIN DECLARE done INT DEFAULT FALSE; DECLARE sale_amount DECIMAL(10,2); DECLARE total DECIMAL(10,2) DEFAULT 0; DECLARE sales_cursor CURSOR FOR SELECT amount FROM sales WHERE status = 'completed'; DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE; OPEN sales_cursor; read_loop: LOOP FETCH sales_cursor INTO sale_amount; IF done THEN LEAVE read_loop; END IF; SET total = total + sale_amount; END LOOP; CLOSE sales_cursor; SELECT total AS total_sales; END;

-- Complex procedure with multiple features
CREATE PROCEDURE ProcessOrders(IN order_date DATE) BEGIN DECLARE order_id INT; DECLARE customer_id INT; DECLARE order_total DECIMAL(10,2); DECLARE done INT DEFAULT FALSE; DECLARE orders_cursor CURSOR FOR SELECT id, customer_id, total FROM orders WHERE DATE(created_at) = order_date; DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE; START TRANSACTION; OPEN orders_cursor; process_loop: LOOP FETCH orders_cursor INTO order_id, customer_id, order_total; IF done THEN LEAVE process_loop; END IF; UPDATE customers SET total_purchases = total_purchases + order_total WHERE id = customer_id; UPDATE orders SET processed = 1 WHERE id = order_id; END LOOP; CLOSE orders_cursor; COMMIT; END;