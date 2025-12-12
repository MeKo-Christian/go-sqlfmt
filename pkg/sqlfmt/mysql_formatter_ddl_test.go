package sqlfmt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMySQLFormatter_DDL_Index(t *testing.T) {
	formatter := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL))

	t.Run("formats CREATE INDEX with USING BTREE", func(t *testing.T) {
		query := "CREATE INDEX idx_user_email ON users (email) USING BTREE;"
		exp := Dedent(`
				CREATE INDEX
				  idx_user_email ON users (email) USING BTREE;
			`)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats CREATE UNIQUE INDEX", func(t *testing.T) {
		query := "CREATE UNIQUE INDEX uk_user_username ON users (username);"
		exp := Dedent(`
				CREATE UNIQUE INDEX
				  uk_user_username ON users (username);
			`)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats CREATE FULLTEXT INDEX", func(t *testing.T) {
		query := "CREATE FULLTEXT INDEX ft_post_content ON posts (title, content);"
		exp := Dedent(`
				CREATE FULLTEXT INDEX
				  ft_post_content ON posts (title, content);
			`)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats CREATE SPATIAL INDEX with USING HASH", func(t *testing.T) {
		query := "CREATE SPATIAL INDEX sp_location ON venues (coordinates) USING HASH;"
		exp := Dedent(`
				CREATE SPATIAL INDEX
				  sp_location ON venues (coordinates) USING HASH;
			`)
		runFormattingTest(t, formatter, query, exp)
	})

	t.Run("formats multi-column index", func(t *testing.T) {
		query := "CREATE INDEX idx_user_status_created ON users (status, created_at);"
		exp := Dedent(`
				CREATE INDEX
				  idx_user_status_created ON users (status, created_at);
			`)
		runFormattingTest(t, formatter, query, exp)
	})
}

func TestMySQLFormatter_DDL_Table(t *testing.T) {
	t.Run("formats ALTER TABLE with ALGORITHM and LOCK", func(t *testing.T) {
		query := "ALTER TABLE users ADD COLUMN age INT, ALGORITHM=INSTANT, LOCK=NONE;"
		exp := Dedent(`
				ALTER TABLE
				  users
				ADD COLUMN
				  age INT,
				  ALGORITHM = INSTANT,
				  LOCK = NONE;
			`)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats complex ALTER TABLE with multiple options", func(t *testing.T) {
		query := "ALTER TABLE products ADD COLUMN discount DECIMAL(5,2) DEFAULT 0.00 AFTER price, " +
			"MODIFY COLUMN name VARCHAR(200) NOT NULL, DROP COLUMN old_field, ALGORITHM=INPLACE, LOCK=SHARED;"
		exp := Dedent(`
				ALTER TABLE
				  products
				ADD COLUMN
				  discount DECIMAL(5, 2) DEFAULT 0.00
				AFTER
				  price,
				MODIFY COLUMN
				  name VARCHAR(200) NOT NULL,
				DROP COLUMN
				  old_field,
				  ALGORITHM = INPLACE,
				  LOCK = SHARED;
			`)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats GENERATED ALWAYS AS VIRTUAL", func(t *testing.T) {
		query := "CREATE TABLE users (id INT, first_name VARCHAR(50), last_name VARCHAR(50), " +
			"full_name VARCHAR(101) GENERATED ALWAYS AS (CONCAT(first_name, ' ', last_name)) VIRTUAL);"
		exp := Dedent(`
				CREATE TABLE users (
				  id INT,
				  first_name VARCHAR(50),
				  last_name VARCHAR(50),
				  full_name VARCHAR(101) GENERATED ALWAYS AS (CONCAT(first_name, ' ', last_name)) VIRTUAL
				);
			`)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats GENERATED ALWAYS AS STORED", func(t *testing.T) {
		query := "CREATE TABLE products (id INT, price DECIMAL(10,2), discount DECIMAL(5,2), " +
			"final_price DECIMAL(10,2) GENERATED ALWAYS AS (price - (price * discount / 100)) STORED);"
		exp := Dedent(`
				CREATE TABLE products (
				  id INT,
				  price DECIMAL(10, 2),
				  discount DECIMAL(5, 2),
				  final_price DECIMAL(10, 2) GENERATED ALWAYS AS (price - (price * discount / 100)) STORED
				);
			`)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats complex generated column expressions", func(t *testing.T) {
		query := "CREATE TABLE products (id INT, name VARCHAR(100), price DECIMAL(10,2), discount DECIMAL(5,2), " +
			"final_price DECIMAL(10, 2) GENERATED ALWAYS AS (CASE WHEN discount > 0 THEN price - (price * discount / 100) " +
			"ELSE price END) STORED);"
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

	t.Run("comprehensive DDL integration test", func(t *testing.T) {
		query := "CREATE TABLE users (id INT PRIMARY KEY AUTO_INCREMENT, username VARCHAR(50) NOT NULL, " +
			"email VARCHAR(100) NOT NULL, status VARCHAR(20) DEFAULT 'active', " +
			"created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, " +
			"updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, " +
			"full_name VARCHAR(150) GENERATED ALWAYS AS (CONCAT(first_name, ' ', last_name)) VIRTUAL); " +
			"CREATE UNIQUE INDEX uk_username ON users (username) USING BTREE; " +
			"CREATE INDEX idx_status_created ON users (status, created_at); " +
			"ALTER TABLE users ADD COLUMN first_name VARCHAR(50), ADD COLUMN last_name VARCHAR(50), " +
			"ALGORITHM=INSTANT, LOCK=NONE;"
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
				ADD COLUMN
				  first_name VARCHAR(50),
				ADD COLUMN
				  last_name VARCHAR(50),
				  ALGORITHM = INSTANT,
				  LOCK = NONE;
			`)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestMySQLFormatter_DDL_StoredRoutines(t *testing.T) {
	t.Run("formats basic stored procedure", func(t *testing.T) {
		query := "CREATE PROCEDURE get_user(IN user_id INT) BEGIN SELECT * FROM users WHERE id = user_id; END;"
		exp := Dedent(`
				CREATE PROCEDURE
				  get_user(IN user_id INT) BEGIN
				    SELECT
				      *
				    FROM
				      users
				    WHERE
				      id = user_id;
				END;
			`)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats stored procedure with parameters", func(t *testing.T) {
		query := "CREATE PROCEDURE update_user(IN user_id INT, IN new_name VARCHAR(100), OUT old_name VARCHAR(100)) " +
			"BEGIN SELECT name INTO old_name FROM users WHERE id = user_id; " +
			"UPDATE users SET name = new_name WHERE id = user_id; END;"
		exp := Dedent(`
				CREATE PROCEDURE
				  update_user(
				    IN user_id INT,
				    IN new_name VARCHAR(100),
				    OUT old_name VARCHAR(100)
				  ) BEGIN
				    SELECT
				      name INTO old_name
				    FROM
				      users
				    WHERE
				      id = user_id;
				  UPDATE
				    users
				  SET
				    name = new_name
				  WHERE
				    id = user_id;
				END;
			`)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats stored function with RETURNS clause", func(t *testing.T) {
		query := "CREATE FUNCTION calculate_tax(price DECIMAL(10,2)) RETURNS DECIMAL(10,2) BEGIN RETURN price * 0.08; END;"
		exp := Dedent(`
				CREATE FUNCTION
				  calculate_tax(price DECIMAL(10, 2)) RETURNS DECIMAL(10, 2) BEGIN
				    RETURN price * 0.08;
				END;
			`)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats procedure with IF/ELSE statements", func(t *testing.T) {
		query := "CREATE PROCEDURE process_order(IN order_id INT) BEGIN DECLARE status VARCHAR(20); " +
			"SELECT order_status INTO status FROM orders WHERE id = order_id; " +
			"IF status = 'pending' THEN UPDATE orders SET order_status = 'processing' WHERE id = order_id; " +
			"ELSEIF status = 'processing' THEN UPDATE orders SET order_status = 'completed' WHERE id = order_id; " +
			"ELSE SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Invalid order status'; END IF; END;"
		exp := Dedent(`
				CREATE PROCEDURE
				  process_order(IN order_id INT) BEGIN
				    DECLARE status VARCHAR(20);
				  SELECT
				    order_status INTO status
				  FROM
				    orders
				  WHERE
				    id = order_id;
				  IF
				    status = 'pending' THEN
				    UPDATE
				      orders
				    SET
				      order_status = 'processing'
				    WHERE
				      id = order_id;
				  ELSEIF status = 'processing' THEN
				  UPDATE
				    orders
				  SET
				    order_status = 'completed'
				  WHERE
				    id = order_id;
				  ELSE SIGNAL SQLSTATE '45000'
				  SET
				    MESSAGE_TEXT = 'Invalid order status';
				  END IF;
				END;
			`)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats procedure with WHILE loop", func(t *testing.T) {
		query := "CREATE PROCEDURE countdown(IN start_value INT) BEGIN DECLARE counter INT DEFAULT start_value; " +
			"WHILE counter > 0 DO SELECT counter; SET counter = counter - 1; END WHILE; END;"
		exp := Dedent(`
				CREATE PROCEDURE
				  countdown(IN start_value INT) BEGIN
				    DECLARE counter INT DEFAULT start_value;
				WHILE
				    counter > 0
				    DO
				    SELECT
				      counter;
				  SET
				    counter = counter - 1;
				END WHILE;

				END;
			`)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats procedure with LOOP and LEAVE", func(t *testing.T) {
		query := "CREATE PROCEDURE find_max(IN limit_val INT) BEGIN " +
			"DECLARE i INT DEFAULT 1; DECLARE max_val INT DEFAULT 0; " +
			"my_loop: LOOP IF i > limit_val THEN LEAVE my_loop; END IF; IF i > max_val THEN SET max_val = i; END IF; " +
			"SET i = i + 1; END LOOP my_loop; SELECT max_val; END;"
		exp := Dedent(`
				CREATE PROCEDURE
				  find_max(IN limit_val INT) BEGIN
				    DECLARE i INT DEFAULT 1;
				  DECLARE max_val INT DEFAULT 0;
				my_loop: LOOP
				    IF
				      i > limit_val THEN
				      LEAVE my_loop;
				  END IF;
				  IF
				    i > max_val THEN
				    SET
				      max_val = i;
				  END IF;
				  SET
				    i = i + 1;
				END LOOP my_loop;

				SELECT
				  max_val;

				END;
			`)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats procedure with REPEAT UNTIL", func(t *testing.T) {
		query := "CREATE PROCEDURE repeat_example() BEGIN DECLARE counter INT DEFAULT 0; " +
			"REPEAT SET counter = counter + 1; SELECT CONCAT('Count: ', counter); UNTIL counter >= 5 END REPEAT; END;"
		exp := Dedent(`
				CREATE PROCEDURE
				  repeat_example() BEGIN
				    DECLARE counter INT DEFAULT 0;
				REPEAT
				    SET
				      counter = counter + 1;
				  SELECT
				    CONCAT('Count: ', counter);
				  UNTIL counter >= 5
				END REPEAT;

				END;
			`)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})

	t.Run("formats function with DETERMINISTIC", func(t *testing.T) {
		query := "CREATE FUNCTION get_user_count() RETURNS INT DETERMINISTIC " +
			"BEGIN DECLARE user_count INT; SELECT COUNT(*) INTO user_count FROM users; RETURN user_count; END;"
		exp := Dedent(`
				CREATE FUNCTION
				  get_user_count() RETURNS INT DETERMINISTIC BEGIN
				    DECLARE user_count INT;
				  SELECT
				    COUNT(*) INTO user_count
				  FROM
				    users;
				  RETURN user_count;
				END;
			`)
		result := NewMySQLFormatter(NewDefaultConfig().WithLang(MySQL)).Format(query)
		exp = strings.TrimSpace(strings.ReplaceAll(exp, "\t", DefaultIndent))
		require.Equal(t, exp, result)
	})
}

func TestMySQLFormatter_Format_DDL(t *testing.T) {
	// Keep only the original simple tests that don't fit elsewhere
}
