-- MySQL Stored Procedures: E-commerce order processing
-- Production-style procedures with transaction handling and error management

DELIMITER //

-- Create order processing log table
CREATE TABLE IF NOT EXISTS order_processing_log (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT,
    action VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'success',
    message TEXT,
    processed_by INT,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_order_id (order_id),
    INDEX idx_action (action),
    INDEX idx_processed_at (processed_at)
)//

-- Procedure to create a new order with inventory validation
CREATE PROCEDURE create_order(
    IN p_customer_id INT,
    IN p_shipping_address_id INT,
    IN p_billing_address_id INT,
    IN p_payment_method_id INT,
    OUT p_order_id INT
)
BEGIN
    DECLARE v_order_total DECIMAL(10,2) DEFAULT 0;
    DECLARE v_tax_amount DECIMAL(10,2) DEFAULT 0;
    DECLARE v_shipping_cost DECIMAL(10,2) DEFAULT 10.00;
    DECLARE v_inventory_available BOOLEAN DEFAULT TRUE;
    DECLARE v_exit_handler BOOLEAN DEFAULT FALSE;

    -- Declare handler for SQL exceptions
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        -- Rollback transaction on error
        ROLLBACK;
        SET p_order_id = NULL;

        -- Log error
        INSERT INTO order_processing_log (
            action, status, message, processed_at
        ) VALUES (
            'CREATE_ORDER',
            'error',
            CONCAT('Failed to create order for customer ', p_customer_id, ': ', 'SQL Exception occurred'),
            NOW()
        );

        RESIGNAL;
    END;

    -- Start transaction
    START TRANSACTION;

    -- Validate customer exists
    IF NOT EXISTS (SELECT 1 FROM customers WHERE id = p_customer_id AND is_active = TRUE) THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Invalid or inactive customer';
    END IF;

    -- Create order header
    INSERT INTO orders (
        customer_id,
        order_date,
        status,
        shipping_address_id,
        billing_address_id,
        payment_method_id,
        subtotal,
        tax_amount,
        shipping_cost,
        total_amount
    ) VALUES (
        p_customer_id,
        NOW(),
        'pending',
        p_shipping_address_id,
        p_billing_address_id,
        p_payment_method_id,
        0, -- Will be updated
        0, -- Will be updated
        v_shipping_cost,
        0  -- Will be updated
    );

    SET p_order_id = LAST_INSERT_ID();

    -- Log order creation
    INSERT INTO order_processing_log (
        order_id, action, message, processed_at
    ) VALUES (
        p_order_id,
        'ORDER_CREATED',
        CONCAT('Order created for customer ', p_customer_id),
        NOW()
    );

    -- Commit transaction
    COMMIT;

EXCEPTION
    WHEN SQLSTATE '45000' THEN
        -- Re-raise validation errors
        RESIGNAL;
END//

-- Procedure to add item to order with inventory check
CREATE PROCEDURE add_order_item(
    IN p_order_id INT,
    IN p_product_id INT,
    IN p_variant_id INT,
    IN p_quantity INT,
    IN p_unit_price DECIMAL(10,2)
)
BEGIN
    DECLARE v_available_quantity INT DEFAULT 0;
    DECLARE v_current_inventory INT DEFAULT 0;
    DECLARE v_inventory_policy ENUM('deny', 'continue');
    DECLARE v_product_name VARCHAR(255);
    DECLARE v_exit_handler BOOLEAN DEFAULT FALSE;

    -- Declare handler for SQL exceptions
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        INSERT INTO order_processing_log (
            order_id, action, status, message, processed_at
        ) VALUES (
            p_order_id,
            'ADD_ITEM',
            'error',
            CONCAT('Failed to add item ', p_product_id, ' to order: SQL Exception'),
            NOW()
        );
        RESIGNAL;
    END;

    -- Start transaction
    START TRANSACTION;

    -- Validate order exists and is in pending status
    IF NOT EXISTS (SELECT 1 FROM orders WHERE id = p_order_id AND status = 'pending') THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Order not found or not in pending status';
    END IF;

    -- Get product information and inventory
    SELECT
        p.name,
        COALESCE(pv.inventory_quantity, p.inventory_quantity) as inventory_qty,
        COALESCE(pv.inventory_policy, p.inventory_policy) as inventory_policy
    INTO
        v_product_name,
        v_current_inventory,
        v_inventory_policy
    FROM products p
    LEFT JOIN product_variants pv ON pv.id = p_variant_id AND pv.product_id = p.id
    WHERE p.id = p_product_id;

    IF v_product_name IS NULL THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Product not found';
    END IF;

    -- Check inventory availability
    SET v_available_quantity = v_current_inventory;

    IF v_inventory_policy = 'deny' AND v_available_quantity < p_quantity THEN
        SIGNAL SQLSTATE '45000'
        SET MESSAGE_TEXT = CONCAT('Insufficient inventory for product: ', v_product_name);
    END IF;

    -- Add order item
    INSERT INTO order_items (
        order_id,
        product_id,
        variant_id,
        product_name,
        quantity,
        unit_price,
        total_price
    ) VALUES (
        p_order_id,
        p_product_id,
        p_variant_id,
        v_product_name,
        p_quantity,
        p_unit_price,
        p_quantity * p_unit_price
    );

    -- Update inventory if policy allows
    IF v_inventory_policy = 'continue' OR v_available_quantity >= p_quantity THEN
        IF p_variant_id IS NOT NULL THEN
            UPDATE product_variants
            SET inventory_quantity = inventory_quantity - p_quantity
            WHERE id = p_variant_id;
        ELSE
            UPDATE products
            SET inventory_quantity = inventory_quantity - p_quantity
            WHERE id = p_product_id;
        END IF;
    END IF;

    -- Recalculate order totals
    CALL recalculate_order_totals(p_order_id);

    -- Log successful addition
    INSERT INTO order_processing_log (
        order_id, action, message, processed_at
    ) VALUES (
        p_order_id,
        'ITEM_ADDED',
        CONCAT('Added ', p_quantity, ' x ', v_product_name, ' to order'),
        NOW()
    );

    COMMIT;

EXCEPTION
    WHEN SQLSTATE '45000' THEN
        RESIGNAL;
END//

-- Procedure to recalculate order totals
CREATE PROCEDURE recalculate_order_totals(IN p_order_id INT)
BEGIN
    DECLARE v_subtotal DECIMAL(10,2) DEFAULT 0;
    DECLARE v_tax_rate DECIMAL(5,4) DEFAULT 0.08; -- 8% tax rate
    DECLARE v_tax_amount DECIMAL(10,2) DEFAULT 0;
    DECLARE v_shipping_cost DECIMAL(10,2) DEFAULT 10.00;
    DECLARE v_total DECIMAL(10,2) DEFAULT 0;

    -- Calculate subtotal from order items
    SELECT COALESCE(SUM(total_price), 0)
    INTO v_subtotal
    FROM order_items
    WHERE order_id = p_order_id;

    -- Calculate tax
    SET v_tax_amount = v_subtotal * v_tax_rate;

    -- Calculate total
    SET v_total = v_subtotal + v_tax_amount + v_shipping_cost;

    -- Update order
    UPDATE orders
    SET
        subtotal = v_subtotal,
        tax_amount = v_tax_amount,
        total_amount = v_total,
        updated_at = NOW()
    WHERE id = p_order_id;

END//

-- Function to get order summary
CREATE FUNCTION get_order_summary(p_order_id INT)
RETURNS JSON
DETERMINISTIC
READS SQL DATA
BEGIN
    DECLARE v_result JSON;

    SELECT JSON_OBJECT(
        'order_id', o.id,
        'customer_id', o.customer_id,
        'status', o.status,
        'order_date', o.order_date,
        'subtotal', o.subtotal,
        'tax_amount', o.tax_amount,
        'shipping_cost', o.shipping_cost,
        'total_amount', o.total_amount,
        'items', (
            SELECT JSON_ARRAYAGG(
                JSON_OBJECT(
                    'product_id', oi.product_id,
                    'product_name', oi.product_name,
                    'quantity', oi.quantity,
                    'unit_price', oi.unit_price,
                    'total_price', oi.total_price
                )
            )
            FROM order_items oi
            WHERE oi.order_id = o.id
        )
    ) INTO v_result
    FROM orders o
    WHERE o.id = p_order_id;

    RETURN v_result;
END//

-- Procedure to process payment and complete order
CREATE PROCEDURE process_payment(
    IN p_order_id INT,
    IN p_payment_amount DECIMAL(10,2),
    IN p_processed_by INT
)
BEGIN
    DECLARE v_order_total DECIMAL(10,2);
    DECLARE v_current_status VARCHAR(20);
    DECLARE v_payment_success BOOLEAN DEFAULT FALSE;

    -- Declare handler for SQL exceptions
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        INSERT INTO order_processing_log (
            order_id, action, status, message, processed_by, processed_at
        ) VALUES (
            p_order_id,
            'PAYMENT_PROCESSING',
            'error',
            'Payment processing failed due to system error',
            p_processed_by,
            NOW()
        );
        RESIGNAL;
    END;

    START TRANSACTION;

    -- Get order information
    SELECT total_amount, status
    INTO v_order_total, v_current_status
    FROM orders
    WHERE id = p_order_id;

    IF v_current_status != 'pending' THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Order is not in pending status';
    END IF;

    -- Validate payment amount
    IF ABS(p_payment_amount - v_order_total) > 0.01 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Payment amount does not match order total';
    END IF;

    -- Simulate payment processing (would integrate with payment gateway)
    SET v_payment_success = TRUE; -- Assume success for demo

    IF v_payment_success THEN
        -- Update order status
        UPDATE orders
        SET
            status = 'paid',
            paid_at = NOW(),
            updated_at = NOW()
        WHERE id = p_order_id;

        -- Log successful payment
        INSERT INTO order_processing_log (
            order_id, action, status, message, processed_by, processed_at
        ) VALUES (
            p_order_id,
            'PAYMENT_COMPLETED',
            'success',
            CONCAT('Payment of $', p_payment_amount, ' processed successfully'),
            p_processed_by,
            NOW()
        );

        COMMIT;
    ELSE
        -- Payment failed
        INSERT INTO order_processing_log (
            order_id, action, status, message, processed_by, processed_at
        ) VALUES (
            p_order_id,
            'PAYMENT_FAILED',
            'error',
            'Payment processing failed',
            p_processed_by,
            NOW()
        );

        ROLLBACK;
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Payment processing failed';
    END IF;

EXCEPTION
    WHEN SQLSTATE '45000' THEN
        RESIGNAL;
END//

DELIMITER ;