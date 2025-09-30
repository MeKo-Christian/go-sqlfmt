-- Complex stored procedures with advanced control flow structures
DELIMITER //

CREATE PROCEDURE ProcessMonthlyReport(IN report_month DATE, IN report_year INT)
BEGIN
    DECLARE current_dept_id INT;
    DECLARE dept_done BOOLEAN DEFAULT FALSE;
    DECLARE total_employees INT DEFAULT 0;
    DECLARE total_salary DECIMAL(15,2) DEFAULT 0;
    DECLARE avg_performance DECIMAL(5,2) DEFAULT 0;

    -- Cursor for departments
    DECLARE dept_cursor CURSOR FOR
        SELECT id FROM departments WHERE active = 1 ORDER BY name;

    DECLARE CONTINUE HANDLER FOR NOT FOUND SET dept_done = TRUE;

    -- Temporary table for results
    CREATE TEMPORARY TABLE monthly_report (
        department_id INT,
        department_name VARCHAR(100),
        employee_count INT,
        total_salary DECIMAL(15,2),
        avg_salary DECIMAL(12,2),
        top_performer VARCHAR(100),
        performance_score DECIMAL(5,2),
        budget_utilization DECIMAL(5,2)
    );

    OPEN dept_cursor;

    dept_loop: LOOP
        FETCH dept_cursor INTO current_dept_id;

        IF dept_done THEN
            LEAVE dept_loop;
        END IF;

        -- Calculate department metrics
        SELECT
            COUNT(e.id),
            COALESCE(SUM(e.salary), 0),
            COALESCE(AVG(e.salary), 0),
            COALESCE(AVG(p.score), 0)
        INTO
            total_employees,
            total_salary,
            @avg_salary,
            avg_performance
        FROM employees e
        LEFT JOIN performance_reviews p ON e.id = p.employee_id
            AND YEAR(p.review_date) = report_year
            AND MONTH(p.review_date) = MONTH(report_month)
        WHERE e.department_id = current_dept_id
            AND e.active = 1;

        -- Get top performer
        SELECT COALESCE(e.name, 'N/A') INTO @top_performer
        FROM employees e
        JOIN performance_reviews p ON e.id = p.employee_id
        WHERE e.department_id = current_dept_id
            AND YEAR(p.review_date) = report_year
            AND MONTH(p.review_date) = MONTH(report_month)
        ORDER BY p.score DESC
        LIMIT 1;

        -- Calculate budget utilization
        SELECT COALESCE(
            (total_salary / NULLIF(d.budget, 0)) * 100, 0
        ) INTO @budget_util
        FROM departments d
        WHERE d.id = current_dept_id;

        -- Insert into temporary table
        INSERT INTO monthly_report (
            department_id,
            department_name,
            employee_count,
            total_salary,
            avg_salary,
            top_performer,
            performance_score,
            budget_utilization
        )
        SELECT
            d.id,
            d.name,
            total_employees,
            total_salary,
            @avg_salary,
            @top_performer,
            avg_performance,
            @budget_util
        FROM departments d
        WHERE d.id = current_dept_id;

    END LOOP dept_loop;

    CLOSE dept_cursor;

    -- Return results
    SELECT * FROM monthly_report ORDER BY total_salary DESC;

    -- Cleanup
    DROP TEMPORARY TABLE monthly_report;

END //

DELIMITER ;

-- Procedure with nested cursors and complex error handling
DELIMITER //

CREATE PROCEDURE ProcessBulkUpdates(
    IN batch_size INT,
    OUT processed_count INT,
    OUT error_count INT
)
BEGIN
    DECLARE batch_done BOOLEAN DEFAULT FALSE;
    DECLARE record_done BOOLEAN DEFAULT FALSE;
    DECLARE current_record_id INT;
    DECLARE current_operation VARCHAR(50);
    DECLARE exit_handler BOOLEAN DEFAULT FALSE;

    DECLARE batch_cursor CURSOR FOR
        SELECT id, operation_type
        FROM pending_updates
        WHERE status = 'pending'
        ORDER BY priority DESC, created_at ASC
        LIMIT batch_size;

    DECLARE record_cursor CURSOR FOR
        SELECT id FROM update_queue
        WHERE batch_id = current_record_id
        ORDER BY sequence_number;

    DECLARE CONTINUE HANDLER FOR NOT FOUND SET record_done = TRUE;
    DECLARE CONTINUE HANDLER FOR SQLEXCEPTION BEGIN
        SET error_count = error_count + 1;
        ROLLBACK;
        START TRANSACTION;
    END;

    SET processed_count = 0;
    SET error_count = 0;

    START TRANSACTION;

    OPEN batch_cursor;

    batch_loop: LOOP
        FETCH batch_cursor INTO current_record_id, current_operation;

        IF batch_done THEN
            LEAVE batch_loop;
        END IF;

        SET record_done = FALSE;

        -- Process records within this batch
        OPEN record_cursor;

        record_loop: LOOP
            FETCH record_cursor INTO @record_id;

            IF record_done THEN
                LEAVE record_loop;
            END IF;

            -- Execute operation based on type
            CASE current_operation
                WHEN 'UPDATE_SALARY' THEN
                    UPDATE employees
                    SET salary = salary * 1.05
                    WHERE id = @record_id;
                WHEN 'UPDATE_STATUS' THEN
                    UPDATE employees
                    SET status = 'active'
                    WHERE id = @record_id;
                WHEN 'DELETE_RECORD' THEN
                    DELETE FROM employees
                    WHERE id = @record_id;
                ELSE
                    SIGNAL SQLSTATE '45000'
                    SET MESSAGE_TEXT = 'Unknown operation type';
            END CASE;

            SET processed_count = processed_count + 1;

        END LOOP record_loop;

        CLOSE record_cursor;

        -- Mark batch as processed
        UPDATE pending_updates
        SET status = 'completed', processed_at = NOW()
        WHERE id = current_record_id;

    END LOOP batch_loop;

    CLOSE batch_cursor;

    COMMIT;

END //

DELIMITER ;

-- Procedure with WHILE loops and conditional logic
DELIMITER //

CREATE PROCEDURE GenerateSalaryReport(
    IN start_date DATE,
    IN end_date DATE,
    IN department_filter VARCHAR(100)
)
BEGIN
    DECLARE current_month DATE DEFAULT start_date;
    DECLARE report_data JSON DEFAULT JSON_OBJECT();
    DECLARE month_key VARCHAR(7);
    DECLARE dept_condition VARCHAR(200) DEFAULT '';

    -- Build department filter condition
    IF department_filter IS NOT NULL AND department_filter != '' THEN
        SET dept_condition = CONCAT(' AND d.name LIKE "%', department_filter, '%"');
    END IF;

    -- Create temporary table for results
    CREATE TEMPORARY TABLE salary_trends (
        report_month VARCHAR(7) PRIMARY KEY,
        total_employees INT DEFAULT 0,
        total_salary DECIMAL(15,2) DEFAULT 0,
        avg_salary DECIMAL(12,2) DEFAULT 0,
        min_salary DECIMAL(10,2) DEFAULT 0,
        max_salary DECIMAL(10,2) DEFAULT 0,
        department_breakdown JSON
    );

    -- Loop through months
    WHILE current_month <= end_date DO
        SET month_key = DATE_FORMAT(current_month, '%Y-%m');

        -- Calculate monthly statistics
        SET @sql = CONCAT('
            INSERT INTO salary_trends (
                report_month, total_employees, total_salary,
                avg_salary, min_salary, max_salary, department_breakdown
            )
            SELECT
                ?, COUNT(e.id), SUM(e.salary), AVG(e.salary),
                MIN(e.salary), MAX(e.salary),
                JSON_OBJECTAGG(d.name, JSON_OBJECT(
                    "count", COUNT(e.id),
                    "total", SUM(e.salary),
                    "avg", AVG(e.salary)
                ))
            FROM employees e
            JOIN departments d ON e.department_id = d.id
            WHERE e.hire_date <= LAST_DAY(?)
                AND (e.termination_date IS NULL OR e.termination_date >= ?)
                AND e.active = 1',
            dept_condition,
            ' GROUP BY d.id, d.name'
        );

        PREPARE stmt FROM @sql;
        EXECUTE stmt USING month_key, current_month, current_month;
        DEALLOCATE PREPARE stmt;

        -- Move to next month
        SET current_month = DATE_ADD(current_month, INTERVAL 1 MONTH);
    END WHILE;

    -- Generate final report
    SELECT
        report_month,
        total_employees,
        total_salary,
        avg_salary,
        min_salary,
        max_salary,
        department_breakdown
    FROM salary_trends
    ORDER BY report_month;

    -- Cleanup
    DROP TEMPORARY TABLE salary_trends;

END //

DELIMITER ;

-- Procedure with REPEAT loops and complex validation
DELIMITER //

CREATE PROCEDURE ValidateAndProcessOrders(
    IN max_retries INT DEFAULT 3,
    OUT success_count INT,
    OUT failure_count INT
)
BEGIN
    DECLARE order_id INT;
    DECLARE validation_passed BOOLEAN DEFAULT FALSE;
    DECLARE retry_count INT DEFAULT 0;
    DECLARE done BOOLEAN DEFAULT FALSE;

    DECLARE order_cursor CURSOR FOR
        SELECT id FROM orders
        WHERE status = 'pending_validation'
        ORDER BY created_at ASC;

    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE;

    SET success_count = 0;
    SET failure_count = 0;

    OPEN order_cursor;

    order_loop: LOOP
        FETCH order_cursor INTO order_id;

        IF done THEN
            LEAVE order_loop;
        END IF;

        SET validation_passed = FALSE;
        SET retry_count = 0;

        -- Retry validation up to max_retries times
        REPEAT
            SET retry_count = retry_count + 1;

            -- Perform validation checks
            IF EXISTS(
                SELECT 1 FROM order_items oi
                JOIN products p ON oi.product_id = p.id
                WHERE oi.order_id = order_id
                    AND p.stock_quantity < oi.quantity
            ) THEN
                -- Insufficient stock
                UPDATE orders SET
                    status = 'validation_failed',
                    failure_reason = 'Insufficient stock',
                    validated_at = NOW()
                WHERE id = order_id;
                LEAVE order_loop;
            END IF;

            IF NOT EXISTS(
                SELECT 1 FROM customers c
                WHERE c.id = (SELECT customer_id FROM orders WHERE id = order_id)
                    AND c.credit_limit >= (
                        SELECT SUM(oi.quantity * p.price)
                        FROM order_items oi
                        JOIN products p ON oi.product_id = p.id
                        WHERE oi.order_id = order_id
                    )
            ) THEN
                -- Insufficient credit
                UPDATE orders SET
                    status = 'validation_failed',
                    failure_reason = 'Insufficient credit limit',
                    validated_at = NOW()
                WHERE id = order_id;
                LEAVE order_loop;
            END IF;

            -- All validations passed
            SET validation_passed = TRUE;

        UNTIL validation_passed = TRUE OR retry_count >= max_retries
        END REPEAT;

        IF validation_passed THEN
            -- Process the order
            UPDATE orders SET
                status = 'validated',
                validated_at = NOW()
            WHERE id = order_id;

            SET success_count = success_count + 1;
        ELSE
            -- Max retries exceeded
            UPDATE orders SET
                status = 'validation_failed',
                failure_reason = 'Validation retry limit exceeded',
                validated_at = NOW()
            WHERE id = order_id;

            SET failure_count = failure_count + 1;
        END IF;

    END LOOP order_loop;

    CLOSE order_cursor;

END //

DELIMITER ;</content>
<parameter name="filePath">/mnt/projekte/Code/go-sqlfmt/testdata/golden/mysql/complex_stored_procedures.sql