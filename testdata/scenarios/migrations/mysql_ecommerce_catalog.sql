-- MySQL Migration: E-commerce product catalog system
-- Version: 002
-- Description: Product catalog with categories, inventory, and pricing

-- Create product_categories table
CREATE TABLE product_categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id INT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    sort_order INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES product_categories(id) ON DELETE SET NULL,
    INDEX idx_parent_id (parent_id),
    INDEX idx_slug (slug),
    INDEX idx_active (is_active),
    INDEX idx_sort_order (sort_order)
);

-- Create products table
CREATE TABLE products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    category_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sku VARCHAR(50) UNIQUE NOT NULL,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    compare_at_price DECIMAL(10,2) NULL CHECK (compare_at_price >= price),
    cost_price DECIMAL(10,2) NULL,
    inventory_quantity INT DEFAULT 0 CHECK (inventory_quantity >= 0),
    inventory_policy ENUM('deny', 'continue') DEFAULT 'deny',
    weight DECIMAL(8,3) NULL,
    weight_unit ENUM('kg', 'g', 'lb', 'oz') DEFAULT 'kg',
    requires_shipping BOOLEAN DEFAULT TRUE,
    is_active BOOLEAN DEFAULT TRUE,
    is_featured BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES product_categories(id),
    INDEX idx_category_id (category_id),
    INDEX idx_sku (sku),
    INDEX idx_active (is_active),
    INDEX idx_featured (is_featured),
    INDEX idx_price (price),
    FULLTEXT idx_search (name, description)
);

-- Create product_images table
CREATE TABLE product_images (
    id INT AUTO_INCREMENT PRIMARY KEY,
    product_id INT NOT NULL,
    image_url VARCHAR(500) NOT NULL,
    alt_text VARCHAR(255),
    sort_order INT DEFAULT 0,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    INDEX idx_product_id (product_id),
    INDEX idx_primary (is_primary)
);

-- Create product_variants table for size/color variations
CREATE TABLE product_variants (
    id INT AUTO_INCREMENT PRIMARY KEY,
    product_id INT NOT NULL,
    sku_suffix VARCHAR(20),
    option1_name VARCHAR(50), -- e.g., "Size"
    option1_value VARCHAR(50), -- e.g., "Large"
    option2_name VARCHAR(50), -- e.g., "Color"
    option2_value VARCHAR(50), -- e.g., "Blue"
    option3_name VARCHAR(50), -- e.g., "Material"
    option3_value VARCHAR(50), -- e.g., "Cotton"
    price_modifier DECIMAL(10,2) DEFAULT 0,
    inventory_quantity INT DEFAULT 0 CHECK (inventory_quantity >= 0),
    weight_modifier DECIMAL(8,3) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    INDEX idx_product_id (product_id),
    INDEX idx_active (is_active),
    UNIQUE KEY unique_variant_options (product_id, option1_value, option2_value, option3_value)
);

-- Insert sample categories
INSERT INTO product_categories (name, description, slug, sort_order) VALUES
('Electronics', 'Electronic devices and accessories', 'electronics', 1),
('Clothing', 'Fashion and apparel', 'clothing', 2),
('Home & Garden', 'Home improvement and garden supplies', 'home-garden', 3),
('Sports & Outdoors', 'Sports equipment and outdoor gear', 'sports-outdoors', 4);

-- Insert sample product
INSERT INTO products (category_id, name, description, sku, price, inventory_quantity, weight, weight_unit) VALUES
(1, 'Wireless Bluetooth Headphones', 'High-quality wireless headphones with noise cancellation', 'WBH-001', 199.99, 50, 0.3, 'kg');

-- Insert product images
INSERT INTO product_images (product_id, image_url, alt_text, sort_order, is_primary) VALUES
(1, '/images/products/wbh-001-main.jpg', 'Wireless Bluetooth Headphones - Main View', 1, TRUE),
(1, '/images/products/wbh-001-side.jpg', 'Wireless Bluetooth Headphones - Side View', 2, FALSE);

-- Insert product variants
INSERT INTO product_variants (product_id, option1_name, option1_value, inventory_quantity) VALUES
(1, 'Color', 'Black', 25),
(1, 'Color', 'White', 25);