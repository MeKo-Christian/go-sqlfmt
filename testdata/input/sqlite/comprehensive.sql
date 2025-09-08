-- SQLite Comprehensive Test File
-- This file demonstrates all major SQLite features supported by the formatter

-- Comments: SQLite supports -- comments but not # comments
-- PRAGMA statements (top-level, minimal formatting)
PRAGMA foreign_keys = ON;PRAGMA journal_mode = WAL;PRAGMA synchronous = NORMAL;

-- DDL with generated columns and STRICT
CREATE TABLE orders (id INTEGER PRIMARY KEY AUTOINCREMENT,customer_id INTEGER NOT NULL,subtotal REAL NOT NULL,tax_rate REAL DEFAULT 0.08,total REAL GENERATED ALWAYS AS (subtotal + (subtotal * tax_rate)) STORED,status TEXT DEFAULT 'pending',created_at TEXT DEFAULT (datetime('now','localtime')),metadata TEXT) STRICT;

-- DDL with various identifier quoting styles (double quotes, backticks, brackets)
CREATE TABLE "user profiles" (user_id INTEGER,[full name] TEXT,`profile_data` TEXT,settings TEXT);

-- Index creation with IF NOT EXISTS
CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_customer_date ON orders(customer_id, date(created_at));CREATE INDEX IF NOT EXISTS idx_profiles_name ON "user profiles"([full name]);

-- WITHOUT ROWID table
CREATE TABLE settings_lookup(key TEXT PRIMARY KEY,value TEXT,category TEXT) WITHOUT ROWID;

-- Trigger with complex body
CREATE TRIGGER update_order_timestamp AFTER UPDATE ON orders FOR EACH ROW WHEN NEW.status != OLD.status BEGIN UPDATE orders SET modified_at = datetime('now') WHERE id = NEW.id; INSERT INTO order_audit (order_id, old_status, new_status, changed_at) VALUES (NEW.id, OLD.status, NEW.status, datetime('now')); END;

-- View with CTE and window functions
CREATE VIEW customer_analytics AS WITH monthly_orders AS (SELECT customer_id,strftime('%Y-%m', created_at) as month,COUNT(*) as order_count,SUM(total) as monthly_total FROM orders GROUP BY customer_id, strftime('%Y-%m', created_at)),customer_rankings AS (SELECT customer_id,month,order_count,monthly_total,ROW_NUMBER() OVER (PARTITION BY month ORDER BY monthly_total DESC) as monthly_rank,LAG(monthly_total, 1, 0) OVER (PARTITION BY customer_id ORDER BY month) as prev_month_total FROM monthly_orders) SELECT customer_id,month,order_count,monthly_total,monthly_rank,monthly_total - prev_month_total as growth FROM customer_rankings;

-- Complex query with all placeholder types
SELECT o.id,o.customer_id,o.total,up."full name" as customer_name,o.status FROM orders o JOIN "user profiles" up ON o.customer_id = up.user_id WHERE o.id = ? AND o.customer_id = :customer_id AND o.total >= @min_amount AND o.status = $status AND o.created_at > ?2 AND up.`profile_data` IS NOT NULL;

-- JSON operations (SQLite 3.38+)
SELECT customer_id,metadata->>'name' as customer_name,metadata->'preferences'->>'theme' as theme,metadata->'settings' as all_settings FROM orders WHERE metadata->>'status' = 'premium' AND metadata->'preferences'->>'notifications' = 'enabled';

-- LIMIT variations (both SQLite styles)
SELECT * FROM orders ORDER BY total DESC LIMIT 10 OFFSET 20;
SELECT * FROM orders ORDER BY created_at DESC LIMIT 5, 15;

-- String concatenation with || operator
SELECT customer_id || ': ' || status as customer_status,'Order #' || CAST(id AS TEXT) || ' - $' || CAST(total AS TEXT) as order_summary FROM orders WHERE total > 100.0;

-- Blob literals and binary data
INSERT INTO file_storage (filename, content_type, data) VALUES ('document.pdf', 'application/pdf', X'255044462D312E34'),('image.png', 'image/png', X'89504E470D0A1A0A');

-- UPSERT with ON CONFLICT
INSERT INTO customer_stats (customer_id, total_orders, total_spent) VALUES (?1, ?2, ?3) ON CONFLICT(customer_id) DO UPDATE SET total_orders = excluded.total_orders + customer_stats.total_orders,total_spent = excluded.total_spent + customer_stats.total_spent,updated_at = datetime('now');

-- INSERT OR variations
INSERT OR IGNORE INTO temp_cache (key, value) VALUES ('session:123', 'user_data');
INSERT OR REPLACE INTO user_sessions (user_id, session_token, expires_at) VALUES (42, 'abc123def456', datetime('now', '+24 hours'));

-- NULL handling with IS DISTINCT FROM
SELECT * FROM orders WHERE customer_id IS NOT NULL AND status IS DISTINCT FROM 'cancelled' AND total IS NOT DISTINCT FROM 0.0 AND metadata IS NOT NULL;

-- Pattern matching with GLOB and LIKE
SELECT * FROM files WHERE filename GLOB '*.{jpg,png,gif}' OR path LIKE '%/uploads/%' AND filename NOT GLOB 'temp*';

-- Complex CTE with RECURSIVE and window functions
WITH RECURSIVE category_hierarchy AS (SELECT id, name, parent_id, 0 as depth, name as path FROM categories WHERE parent_id IS NULL UNION ALL SELECT c.id, c.name, c.parent_id, ch.depth + 1, ch.path || ' > ' || c.name FROM categories c JOIN category_hierarchy ch ON c.parent_id = ch.id WHERE ch.depth < 10),sales_by_category AS (SELECT ch.id as category_id,ch.name,ch.depth,ch.path,COUNT(DISTINCT oi.order_id) as total_orders,SUM(oi.quantity * oi.unit_price) as total_revenue,AVG(oi.quantity * oi.unit_price) as avg_order_value,RANK() OVER (PARTITION BY ch.depth ORDER BY SUM(oi.quantity * oi.unit_price) DESC) as revenue_rank,LAG(SUM(oi.quantity * oi.unit_price), 1) OVER (PARTITION BY ch.depth ORDER BY SUM(oi.quantity * oi.unit_price) DESC) as prev_category_revenue FROM category_hierarchy ch LEFT JOIN products p ON p.category_id = ch.id LEFT JOIN order_items oi ON oi.product_id = p.id WHERE oi.created_at >= date('now', '-3 months') OR oi.created_at IS NULL GROUP BY ch.id, ch.name, ch.depth, ch.path) SELECT category_id,name,depth,path,total_orders,total_revenue,avg_order_value,revenue_rank,COALESCE(total_revenue - prev_category_revenue, 0) as revenue_gap FROM sales_by_category WHERE total_revenue > 1000 OR total_orders IS NULL ORDER BY depth ASC, revenue_rank ASC;

-- Advanced window functions with frames
SELECT order_id,customer_id,order_date,total,SUM(total) OVER (PARTITION BY customer_id ORDER BY order_date ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) as running_total,AVG(total) OVER (PARTITION BY customer_id ORDER BY order_date ROWS BETWEEN 2 PRECEDING AND 2 FOLLOWING) as moving_avg,LAG(total, 1, 0) OVER (PARTITION BY customer_id ORDER BY order_date) as prev_order_total,LEAD(total) OVER (PARTITION BY customer_id ORDER BY order_date RANGE BETWEEN CURRENT ROW AND UNBOUNDED FOLLOWING) as next_order_total,COUNT(*) OVER (PARTITION BY customer_id) as customer_total_orders FROM orders WHERE order_date >= date('now', '-1 year');

-- Clean up statements
DROP TRIGGER IF EXISTS update_order_timestamp;
DROP VIEW IF EXISTS customer_analytics;
DROP INDEX IF EXISTS idx_orders_customer_date;
DROP TABLE IF EXISTS orders;