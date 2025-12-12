-- Basic function
CREATE FUNCTION
  GetUserCount() RETURNS INT DETERMINISTIC BEGIN
    DECLARE user_count INT;
  SELECT
    COUNT(*) INTO user_count
  FROM
    users;
  RETURN user_count;
END;

-- Function with parameters
CREATE FUNCTION
  CalculateDiscount(price DECIMAL(10, 2), discount_percent INT) RETURNS DECIMAL(10, 2) DETERMINISTIC BEGIN
    RETURN price - (price * discount_percent / 100);
END;

-- Function with characteristics
CREATE FUNCTION
  GetCustomerLevel(customer_id INT) RETURNS VARCHAR(20) READS SQL DATA SQL SECURITY definer BEGIN
    DECLARE total_purchases DECIMAL(10, 2);
  SELECT
    SUM(amount) INTO total_purchases
  FROM
    orders
  WHERE
    customer_id = customer_id;
  IF
    total_purchases > 10000 THEN
    RETURN 'Platinum';
  ELSEIF total_purchases > 5000 THEN
  RETURN 'Gold';
  ELSEIF total_purchases > 1000 THEN
  RETURN 'Silver';
  ELSE
  RETURN 'Bronze';
  END IF;
END;

-- Function with local variables and complex logic
CREATE FUNCTION
  CalculateTax(amount DECIMAL(10, 2), tax_code VARCHAR(10)) RETURNS DECIMAL(10, 2) NOT DETERMINISTIC READS SQL DATA BEGIN
    DECLARE tax_rate DECIMAL(5, 2);
  DECLARE tax_amount DECIMAL(10, 2);
  SELECT
    rate INTO tax_rate
  FROM
    tax_rates
  WHERE
    code = tax_code;
  IF
    tax_rate IS NULL THEN
    SET
      tax_rate = 0.10;
  END IF;
  SET
    tax_amount = amount * tax_rate;
  RETURN tax_amount;
END;

-- Function with nested blocks
CREATE FUNCTION
  ValidateEmail(email VARCHAR(255)) RETURNS BOOLEAN DETERMINISTIC BEGIN
    DECLARE is_valid BOOLEAN DEFAULT FALSE;
BEGIN
    DECLARE at_pos INT;
    DECLARE dot_pos INT;
    SET
      at_pos = LOCATE('@', email);
    SET
      dot_pos = LOCATE('.', email, at_pos);
    IF
      at_pos > 0
      AND dot_pos > at_pos THEN
      SET
        is_valid = TRUE;
    END IF;
  END;
  RETURN is_valid;
END;