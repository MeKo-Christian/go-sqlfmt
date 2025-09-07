-- Phase 8: DDL Essentials Test File for MySQL
-- This file contains comprehensive DDL examples to test MySQL Phase 8 features
-- CREATE INDEX variations
CREATE INDEX
  idx_user_email ON users (email) USING BTREE;

CREATE UNIQUE INDEX
  uk_user_username ON users (username);

CREATE FULLTEXT INDEX
  ft_post_content ON posts (title, content);

CREATE SPATIAL INDEX
  sp_location ON venues (coordinates) USING HASH;

-- Multi-column indexes
CREATE INDEX
  idx_user_status_created ON users (status, created_at, updated_at) USING BTREE;

CREATE UNIQUE INDEX
  uk_product_sku_store ON products (sku, store_id, is_active) USING BTREE;

-- ALTER TABLE with options  
ALTER TABLE
  users
ADD
  COLUMN status VARCHAR(20) DEFAULT 'active',
  ALGORITHM = instant,
  LOCK = none;

ALTER TABLE
  products
MODIFY
  COLUMN price DECIMAL(10, 2) NOT NULL,
  ALGORITHM = inplace,
  LOCK = shared;

ALTER TABLE
  orders
ADD
  CONSTRAINT chk_total CHECK (total > 0),
ADD
  INDEX idx_order_date (created_at),
  ALGORITHM = copy,
  LOCK = exclusive;

-- Generated columns - VIRTUAL
CREATE TABLE orders (
  id INT PRIMARY KEY,
  subtotal DECIMAL(10, 2),
  tax_rate DECIMAL(3, 4),
  tax_amount DECIMAL(10, 2) GENERATED ALWAYS AS (subtotal * tax_rate) VIRTUAL,
  shipping_cost DECIMAL(8, 2) DEFAULT 0.00,
  total_amount DECIMAL(10, 2) GENERATED ALWAYS AS (subtotal + tax_amount + shipping_cost) VIRTUAL
);

-- Generated columns - STORED
CREATE TABLE products (
  id INT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(100) NOT NULL,
  base_price DECIMAL(10, 2) NOT NULL,
  discount_percent DECIMAL(5, 2) DEFAULT 0,
  final_price DECIMAL(10, 2) GENERATED ALWAYS AS (
    CASE
      WHEN discount_percent > 0 THEN base_price - (base_price * discount_percent / 100)
      ELSE base_price
    END
  ) STORED,
  price_category VARCHAR(20) GENERATED ALWAYS AS (
    CASE
      WHEN final_price < 10.00 THEN 'budget'
      WHEN final_price < 50.00 THEN 'standard'
      WHEN final_price < 200.00 THEN 'premium'
      ELSE 'luxury'
    END
  ) VIRTUAL
);

-- Complex DDL integration example
CREATE TABLE users (
  id INT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(50) NOT NULL,
  email VARCHAR(100) NOT NULL,
  first_name VARCHAR(50),
  last_name VARCHAR(50),
  status VARCHAR(20) DEFAULT 'active',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  full_name VARCHAR(150) GENERATED ALWAYS AS (CONCAT(first_name, ' ', last_name)) VIRTUAL,
  username_hash CHAR(64) GENERATED ALWAYS AS (SHA2(LOWER(username), 256)) STORED
);

CREATE UNIQUE INDEX
  uk_username ON users (username) USING BTREE;

CREATE INDEX
  idx_status_created ON users (status, created_at);

CREATE FULLTEXT INDEX
  ft_names ON users (first_name, last_name);

ALTER TABLE
  users
ADD
  COLUMN phone VARCHAR(15),
ADD
  COLUMN is_verified BOOLEAN DEFAULT FALSE,
ADD
  CONSTRAINT chk_username_length CHECK (LENGTH(username) >= 3),
  ALGORITHM = instant,
  LOCK = none;