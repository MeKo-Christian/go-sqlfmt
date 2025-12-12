-- Advanced MySQL stored procedure with nested IF/ELSEIF/ELSE
CREATE PROCEDURE
  CalculateShippingCost(
    IN weight DECIMAL(10, 2),
    IN distance INT,
    OUT cost DECIMAL(10, 2)
  ) BEGIN
    DECLARE base_cost DECIMAL(10, 2);

IF
  weight <= 1 THEN
  SET
    base_cost = 5.00;

ELSEIF weight <= 5 THEN
SET
  base_cost = 10.00;

ELSEIF weight <= 10 THEN
SET
  base_cost = 15.00;

ELSE
SET
  base_cost = 20.00;

END IF;

IF
  distance <= 100 THEN
  SET
    cost = base_cost;

ELSEIF distance <= 500 THEN
SET
  cost = base_cost * 1.5;

ELSE
SET
  cost = base_cost * 2.0;

END IF;

END;

-- MySQL procedure with WHILE loop and nested IF
CREATE PROCEDURE
  ProcessBatch(IN batch_id INT) BEGIN
    DECLARE done INT DEFAULT 0;

DECLARE current_id INT;

DECLARE status VARCHAR(20);

DECLARE item_cursor CURSOR FOR
SELECT
  id
FROM
  batch_items
WHERE
  batch_id = batch_id;

DECLARE CONTINUE HANDLER FOR NOT FOUND
SET
  done = 1;

OPEN item_cursor;

process_loop: WHILE
  done = 0
  DO
  FETCH item_cursor INTO current_id;

IF
  done = 0 THEN
  SELECT
    item_status INTO status
  FROM
    items
  WHERE
    id = current_id;

IF
  status = 'pending' THEN
  UPDATE
    items
  SET
    item_status = 'processed'
  WHERE
    id = current_id;

ELSEIF status = 'error' THEN
UPDATE
  items
SET
  item_status = 'retry'
WHERE
  id = current_id;

END IF;

END IF;

END WHILE;

CLOSE item_cursor;

END;

-- MySQL procedure with REPEAT loop
CREATE PROCEDURE
  GenerateFibonacci(IN n INT) BEGIN
    DECLARE counter INT DEFAULT 0;

DECLARE fib_prev INT DEFAULT 0;

DECLARE fib_curr INT DEFAULT 1;

DECLARE fib_next INT;

CREATE TEMPORARY TABLE IF
  NOT EXISTS fibonacci_results (position INT, value INT);

REPEAT
  INSERT INTO
    fibonacci_results
  VALUES
    (counter, fib_prev);

SET
  fib_next = fib_prev + fib_curr;

SET
  fib_prev = fib_curr;

SET
  fib_curr = fib_next;

SET
  counter = counter + 1;

UNTIL counter >= n
END REPEAT;

SELECT
  *
FROM
  fibonacci_results;

DROP TEMPORARY TABLE fibonacci_results;

END;

-- MySQL procedure with LOOP and LEAVE/ITERATE
CREATE PROCEDURE
  ProcessOrders() BEGIN
    DECLARE v_order_id INT;

DECLARE v_total DECIMAL(10, 2);

DECLARE v_status VARCHAR(20);

DECLARE done INT DEFAULT 0;

DECLARE order_cursor CURSOR FOR
SELECT
  id,
  total,
  status
FROM
  orders
WHERE
  status IN ('pending', 'processing');

DECLARE CONTINUE HANDLER FOR NOT FOUND
SET
  done = 1;

OPEN order_cursor;

order_loop: LOOP
  FETCH order_cursor INTO v_order_id,
  v_total,
  v_status;

IF
  done = 1 THEN
  LEAVE order_loop;

END IF;

IF
  v_status = 'cancelled' THEN
  ITERATE order_loop;

END IF;

IF
  v_total > 1000 THEN
  UPDATE
    orders
  SET
    priority = 'high'
  WHERE
    id = v_order_id;

ELSEIF v_total > 500 THEN
UPDATE
  orders
SET
  priority = 'medium'
WHERE
  id = v_order_id;

ELSE
UPDATE
  orders
SET
  priority = 'low'
WHERE
  id = v_order_id;

END IF;

UPDATE
  orders
SET
  status = 'processed'
WHERE
  id = v_order_id;

END LOOP;

CLOSE order_cursor;

END;

-- Complex MySQL procedure with multiple cursors and nested control flow
CREATE PROCEDURE
  ReconcileInventory(IN warehouse_id INT) BEGIN
    DECLARE v_product_id INT;

DECLARE v_expected_qty INT;

DECLARE v_actual_qty INT;

DECLARE v_difference INT;

DECLARE done INT DEFAULT 0;

DECLARE product_cursor CURSOR FOR
SELECT
  product_id,
  quantity
FROM
  inventory
WHERE
  warehouse_id = warehouse_id;

DECLARE CONTINUE HANDLER FOR NOT FOUND
SET
  done = 0;

START TRANSACTION;

OPEN product_cursor;

reconcile_loop: LOOP
  FETCH product_cursor INTO v_product_id,
  v_expected_qty;

IF
  done THEN
  LEAVE reconcile_loop;

END IF;

SELECT
  COUNT(*) INTO v_actual_qty
FROM
  physical_count
WHERE
  product_id = v_product_id
  AND warehouse_id = warehouse_id;

SET
  v_difference = v_actual_qty - v_expected_qty;

IF
  v_difference != 0 THEN IF
    v_difference > 0 THEN
    INSERT INTO
      inventory_adjustments (
        product_id,
        warehouse_id,
        adjustment_type,
        quantity
      )
    VALUES
      (
        v_product_id,
        warehouse_id,
        'surplus',
        v_difference
      );

ELSEIF v_difference < 0 THEN
INSERT INTO
  inventory_adjustments (
    product_id,
    warehouse_id,
    adjustment_type,
    quantity
  )
VALUES
  (
    v_product_id,
    warehouse_id,
    'shortage',
    ABS(v_difference)
  );

END IF;

UPDATE
  inventory
SET
  quantity = v_actual_qty
WHERE
  product_id = v_product_id
  AND warehouse_id = warehouse_id;

END IF;

END LOOP;

CLOSE product_cursor;

COMMIT;

END;