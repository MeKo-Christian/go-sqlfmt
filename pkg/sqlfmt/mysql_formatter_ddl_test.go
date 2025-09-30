package sqlfmt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMySQLFormatter_Format_DDL(t *testing.T) {
	// Phase 8: DDL Essentials Tests
	t.Run("Phase 8: DDL Essentials", func(t *testing.T) {
		t.Run("formats CREATE INDEX with USING BTREE", func(t *testing.T) {
			query := "CREATE INDEX idx_user_email ON users (email) USING BTREE;"
			exp := Dedent(`
				CREATE INDEX
				  idx_user_email ON users (email) USING BTREE;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats CREATE UNIQUE INDEX", func(t *testing.T) {
			query := "CREATE UNIQUE INDEX uk_user_username ON users (username);"
			exp := Dedent(`
				CREATE UNIQUE INDEX
				  uk_user_username ON users (username);
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats CREATE FULLTEXT INDEX", func(t *testing.T) {
			query := "CREATE FULLTEXT INDEX ft_post_content ON posts (title, content);"
			exp := Dedent(`
				CREATE FULLTEXT INDEX
				  ft_post_content ON posts (title, content);
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats CREATE SPATIAL INDEX with USING HASH", func(t *testing.T) {
			query := "CREATE SPATIAL INDEX sp_location ON venues (coordinates) USING HASH;"
			exp := Dedent(`
				CREATE SPATIAL INDEX
				  sp_location ON venues (coordinates) USING HASH;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats multi-column index", func(t *testing.T) {
			query := "CREATE INDEX idx_user_status_created ON users (status, created_at, updated_at) USING BTREE;"
			exp := Dedent(`
				CREATE INDEX
				  idx_user_status_created ON users (status, created_at, updated_at) USING BTREE;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats ALTER TABLE with ALGORITHM and LOCK", func(t *testing.T) {
			query := "ALTER TABLE users ADD COLUMN status VARCHAR(20) DEFAULT 'active', ALGORITHM=INSTANT, LOCK=NONE;"
			exp := Dedent(`
				ALTER TABLE
				  users
				ADD
				  COLUMN status VARCHAR(20) DEFAULT 'active',
				  ALGORITHM = instant,
				  LOCK = none;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats complex ALTER TABLE with multiple options", func(t *testing.T) {
			query := "ALTER TABLE products MODIFY COLUMN price DECIMAL(10,2) NOT NULL, ADD CONSTRAINT chk_price CHECK (price > 0), " +
				"ALGORITHM=INPLACE, LOCK=SHARED;"
			exp := Dedent(`
				ALTER TABLE
				  products
				MODIFY
				  COLUMN price DECIMAL(10, 2) NOT NULL,
				ADD
				  CONSTRAINT chk_price CHECK (price > 0),
				  ALGORITHM = inplace,
				  LOCK = shared;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats GENERATED ALWAYS AS VIRTUAL", func(t *testing.T) {
			query := "CREATE TABLE orders (id INT PRIMARY KEY, subtotal DECIMAL(10,2), tax_rate DECIMAL(3,4), " +
				"tax_amount DECIMAL(10,2) GENERATED ALWAYS AS (subtotal * tax_rate) VIRTUAL);"
			exp := Dedent(`
				CREATE TABLE orders (
				  id INT PRIMARY KEY,
				  subtotal DECIMAL(10, 2),
				  tax_rate DECIMAL(3, 4),
				  tax_amount DECIMAL(10, 2) GENERATED ALWAYS AS (subtotal * tax_rate) VIRTUAL
				);
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats GENERATED ALWAYS AS STORED", func(t *testing.T) {
			query := "CREATE TABLE orders (id INT PRIMARY KEY, subtotal DECIMAL(10,2), tax_rate DECIMAL(3,4), " +
				"total_amount DECIMAL(10,2) GENERATED ALWAYS AS (subtotal + (subtotal * tax_rate)) STORED);"
			exp := Dedent(`
				CREATE TABLE orders (
				  id INT PRIMARY KEY,
				  subtotal DECIMAL(10, 2),
				  tax_rate DECIMAL(3, 4),
				  total_amount DECIMAL(10, 2) GENERATED ALWAYS AS (subtotal + (subtotal * tax_rate)) STORED
				);
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats complex generated column expressions", func(t *testing.T) {
			query := "CREATE TABLE products (id INT, name VARCHAR(100), price DECIMAL(10,2), discount DECIMAL(5,2), " +
				"final_price DECIMAL(10,2) GENERATED ALWAYS AS (CASE WHEN discount > 0 THEN price - (price * discount / 100) ELSE price END) STORED);"
			exp := Dedent(`
				CREATE TABLE products (
				  id INT,
				  name VARCHAR(100),
				  price DECIMAL(10, 2),
				  discount DECIMAL(5, 2),
				  final_price DECIMAL(10, 2) GENERATED ALWAYS AS (
				    CASE
				      WHEN discount > 0 THEN price - (price * discount / 100)
				      ELSE price
				    END
				  ) STORED
				);
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("comprehensive Phase 8 DDL integration test", func(t *testing.T) {
			query := "CREATE TABLE users (id INT PRIMARY KEY AUTO_INCREMENT, username VARCHAR(50) NOT NULL, email VARCHAR(100) NOT NULL, status VARCHAR(20) DEFAULT 'active', created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, full_name VARCHAR(150) GENERATED ALWAYS AS (CONCAT(first_name, ' ', last_name)) VIRTUAL); CREATE UNIQUE INDEX uk_username ON users (username) USING BTREE; CREATE INDEX idx_status_created ON users (status, created_at); ALTER TABLE users ADD COLUMN first_name VARCHAR(50), ADD COLUMN last_name VARCHAR(50), ALGORITHM=INSTANT, LOCK=NONE;"
			exp := Dedent(`
				CREATE TABLE users (
				  id INT PRIMARY KEY AUTO_INCREMENT,
				  username VARCHAR(50) NOT NULL,
				  email VARCHAR(100) NOT NULL,
				  status VARCHAR(20) DEFAULT 'active',
				  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				  full_name VARCHAR(150) GENERATED ALWAYS AS (CONCAT(first_name, ' ', last_name)) VIRTUAL
				);

				CREATE UNIQUE INDEX
				  uk_username ON users (username) USING BTREE;

				CREATE INDEX
				  idx_status_created ON users (status, created_at);

				ALTER TABLE
				  users
				ADD
				  COLUMN first_name VARCHAR(50),
				ADD
				  COLUMN last_name VARCHAR(50),
				  ALGORITHM = instant,
				  LOCK = none;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})
	})

	t.Run("Phase 9: Stored Routines", func(t *testing.T) {
		t.Run("formats basic stored procedure", func(t *testing.T) {
			query := "CREATE PROCEDURE GetUserCount() BEGIN SELECT COUNT(*) FROM users; END;"
			exp := Dedent(`
				CREATE PROCEDURE
				  GetUserCount() BEGIN
				    SELECT
				      COUNT(*)
				    FROM
				      users;
				
				END;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats stored procedure with parameters", func(t *testing.T) {
			query := "CREATE PROCEDURE UpdateUserStatus(IN user_id INT, IN new_status VARCHAR(20)) BEGIN UPDATE users SET status = new_status WHERE id = user_id; END;"
			exp := Dedent(`
				CREATE PROCEDURE
				  UpdateUserStatus(IN user_id INT, IN new_status VARCHAR(20)) BEGIN
				    UPDATE
				      users
				    SET
				      status = new_status
				    WHERE
				      id = user_id;
				
				END;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats stored function with RETURNS clause", func(t *testing.T) {
			query := "CREATE FUNCTION CalculateDiscount(price DECIMAL(10,2), discount_percent INT) RETURNS DECIMAL(10,2) DETERMINISTIC BEGIN RETURN price - (price * discount_percent / 100); END;"
			exp := Dedent(`
				CREATE FUNCTION
				  CalculateDiscount(price DECIMAL(10, 2), discount_percent INT) RETURNS DECIMAL(10, 2) DETERMINISTIC BEGIN
				    RETURN price - (price * discount_percent / 100);
				
				END;
			`)
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats procedure with IF/ELSE statements", func(t *testing.T) {
			query := "CREATE PROCEDURE CheckStock(IN product_id INT, OUT stock_status VARCHAR(20)) BEGIN DECLARE stock_count INT; SELECT quantity INTO stock_count FROM inventory WHERE id = product_id; IF stock_count > 100 THEN SET stock_status = 'High'; ELSEIF stock_count > 20 THEN SET stock_status = 'Medium'; ELSE SET stock_status = 'Low'; END IF; END;"
			exp := `CREATE PROCEDURE
  CheckStock(IN product_id INT, OUT stock_status VARCHAR(20)) BEGIN
    DECLARE stock_count INT;

SELECT
  quantity INTO stock_count
FROM
  inventory
WHERE
  id = product_id;

IF
  stock_count > 100 THEN
  SET
    stock_status = 'High';

ELSEIF stock_count > 20 THEN
SET
  stock_status = 'Medium';

ELSE
SET
  stock_status = 'Low';

END IF
;

END;`
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats procedure with WHILE loop", func(t *testing.T) {
			query := "CREATE PROCEDURE GenerateNumbers(IN max_num INT) BEGIN DECLARE counter INT DEFAULT 0; CREATE TEMPORARY TABLE numbers (num INT); WHILE counter < max_num DO INSERT INTO numbers VALUES (counter); SET counter = counter + 1; END WHILE; SELECT * FROM numbers; DROP TEMPORARY TABLE numbers; END;"
			exp := `CREATE PROCEDURE
  GenerateNumbers(IN max_num INT) BEGIN
    DECLARE counter INT DEFAULT 0;

CREATE TEMPORARY TABLE numbers (num INT);

WHILE
  counter < max_num
  DO
  INSERT INTO
    numbers
  VALUES
    (counter);

SET
  counter = counter + 1;

END WHILE
;

SELECT
  *
FROM
  numbers;

DROP TEMPORARY TABLE numbers;

END;`
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats procedure with LOOP and LEAVE", func(t *testing.T) {
			query := "CREATE PROCEDURE ProcessBatch(IN batch_size INT) BEGIN DECLARE done INT DEFAULT 0; DECLARE counter INT DEFAULT 0; process_loop: LOOP IF counter >= batch_size THEN LEAVE process_loop; END IF; CALL ProcessSingleItem(counter); SET counter = counter + 1; END LOOP; END;"
			exp := `CREATE PROCEDURE
  ProcessBatch(IN batch_size INT) BEGIN
    DECLARE done INT DEFAULT 0;

DECLARE counter INT DEFAULT 0;

process_loop: LOOP
  IF
    counter >= batch_size THEN
    LEAVE process_loop;

END IF
;

CALL ProcessSingleItem(counter);

SET
  counter = counter + 1;

END LOOP
;

END;`
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats procedure with REPEAT UNTIL", func(t *testing.T) {
			query := "CREATE PROCEDURE WaitForCondition() BEGIN DECLARE attempts INT DEFAULT 0; REPEAT SET attempts = attempts + 1; CALL CheckCondition(@result); UNTIL @result = 1 OR attempts > 10 END REPEAT; END;"
			exp := `CREATE PROCEDURE
  WaitForCondition() BEGIN
    DECLARE attempts INT DEFAULT 0;

REPEAT
  SET
    attempts = attempts + 1;

CALL CheckCondition(@result);

UNTIL @result = 1
OR attempts > 10
END REPEAT
;

END;`
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats function with characteristics", func(t *testing.T) {
			query := "CREATE FUNCTION GetUserName(user_id INT) RETURNS VARCHAR(100) READS SQL DATA SQL SECURITY DEFINER BEGIN DECLARE user_name VARCHAR(100); SELECT name INTO user_name FROM users WHERE id = user_id; RETURN user_name; END;"
			exp := `CREATE FUNCTION
  GetUserName(user_id INT) RETURNS VARCHAR(100) READS SQL DATA SQL SECURITY definer BEGIN
    DECLARE user_name VARCHAR(100);

SELECT
  name INTO user_name
FROM
  users
WHERE
  id = user_id;

RETURN user_name;

END;`
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats procedure with cursor", func(t *testing.T) {
			query := "CREATE PROCEDURE ProcessAllUsers() BEGIN DECLARE done INT DEFAULT FALSE; DECLARE user_id INT; DECLARE user_cursor CURSOR FOR SELECT id FROM users; DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = TRUE; OPEN user_cursor; read_loop: LOOP FETCH user_cursor INTO user_id; IF done THEN LEAVE read_loop; END IF; CALL ProcessUser(user_id); END LOOP; CLOSE user_cursor; END;"
			exp := `CREATE PROCEDURE
  ProcessAllUsers() BEGIN
    DECLARE done INT DEFAULT FALSE;

DECLARE user_id INT;

DECLARE user_cursor CURSOR FOR
SELECT
  id
FROM
  users;

DECLARE CONTINUE HANDLER FOR NOT FOUND
SET
  done = TRUE;

OPEN user_cursor;

read_loop: LOOP
  FETCH user_cursor INTO user_id;

IF
  done THEN
  LEAVE read_loop;

END IF
;

CALL ProcessUser(user_id);

END LOOP
;

CLOSE user_cursor;

END;`
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("handles DELIMITER statements as pass-through", func(t *testing.T) {
			// Note: In real usage, DELIMITER would change how statements are terminated
			// For formatting purposes, we treat it as a pass-through
			query := "DELIMITER $$ CREATE PROCEDURE TestProc() BEGIN SELECT 1; END$$ DELIMITER ;"
			// The formatter should preserve DELIMITER but still format the procedure
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			// Check that the result contains formatted procedure
			require.Contains(t, result, "CREATE PROCEDURE")
			require.Contains(t, result, "BEGIN")
			require.Contains(t, result, "END")
		})

		t.Run("formats nested BEGIN/END blocks", func(t *testing.T) {
			query := "CREATE PROCEDURE NestedBlocks() BEGIN DECLARE x INT; BEGIN DECLARE y INT; SET y = 10; BEGIN DECLARE z INT; SET z = y * 2; END; END; END;"
			exp := `CREATE PROCEDURE
  NestedBlocks() BEGIN
    DECLARE x INT;

BEGIN
  DECLARE y INT;

SET
  y = 10;

BEGIN
  DECLARE z INT;

SET
  z = y * 2;

END;

END;

END;`
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})

		t.Run("formats complex stored procedure", func(t *testing.T) {
			query := "CREATE PROCEDURE ComplexProc(IN category VARCHAR(50), OUT total DECIMAL(10,2)) BEGIN DECLARE done INT DEFAULT 0; DECLARE prod_price DECIMAL(10,2); DECLARE cur CURSOR FOR SELECT price FROM products WHERE category_name = category; DECLARE CONTINUE HANDLER FOR NOT FOUND SET done = 1; SET total = 0; OPEN cur; read_loop: LOOP FETCH cur INTO prod_price; IF done THEN LEAVE read_loop; END IF; SET total = total + prod_price; END LOOP; CLOSE cur; END;"
			exp := `CREATE PROCEDURE
  ComplexProc(IN category VARCHAR(50), OUT total DECIMAL(10, 2)) BEGIN
    DECLARE done INT DEFAULT 0;

DECLARE prod_price DECIMAL(10, 2);

DECLARE cur CURSOR FOR
SELECT
  price
FROM
  products
WHERE
  category_name = category;

DECLARE CONTINUE HANDLER FOR NOT FOUND
SET
  done = 1;

SET
  total = 0;

OPEN cur;

read_loop: LOOP
  FETCH cur INTO prod_price;

IF
  done THEN
  LEAVE read_loop;

END IF
;

SET
  total = total + prod_price;

END LOOP
;

CLOSE cur;

END;`
			result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
			exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
			require.Equal(t, exp, result)
		})
	})
}
