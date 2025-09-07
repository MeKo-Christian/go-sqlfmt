-- MySQL 8.0 Comprehensive Feature Test
-- This file tests all major MySQL features implemented in phases 1-9

-- Phase 1-3: Basic MySQL syntax with comments, identifiers, and placeholders
SELECT /*! SQL_CALC_FOUND_ROWS */ `user_id`, "full_name", 'email_address', 0xFF as hex_flags, 0b1010 as binary_mask, TRUE as is_active
FROM `user_table` # hash comment
WHERE `status` IN (?, ?, ?) -- prepared statement parameters
  AND created_at > ? /* parameter for date filter */
  AND flags & 0b0001 > 0 /*! MySQL version hint */;

-- Phase 4: JSON operators and NULL-safe equality
SELECT 
    u.id,
    profile->'$.name' AS profile_name,
    settings->>'$.theme' AS theme_preference,
    metadata->'$.tags'[0] AS first_tag,
    CASE 
        WHEN preferences->'$.notifications'->>'$.email' = 'enabled' THEN 'EMAIL_ON'
        ELSE 'EMAIL_OFF' 
    END as email_status
FROM users u
LEFT JOIN user_profiles up ON u.id <=> up.user_id  -- NULL-safe join
WHERE u.email REGEXP '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
  AND u.status NOT REGEXP '^(banned|suspended|inactive)$'
  AND (u.flags | 0x10) > 0  -- bitwise operations
  AND u.score << 2 > 100;   -- bitwise shift

-- Phase 5: LIMIT variations and locking
SELECT order_id, customer_name, total_amount, status
FROM orders
WHERE status = 'pending'
  AND created_at >= CURDATE()
ORDER BY total_amount DESC, created_at DESC
LIMIT 20, 10  -- MySQL-style offset, limit
FOR UPDATE;   -- Row locking

-- Phase 6: MySQL upsert with ON DUPLICATE KEY UPDATE
INSERT INTO user_analytics (
    user_id, 
    session_data, 
    page_views, 
    last_visit, 
    browser_info
) VALUES 
    (?, '{"page": "dashboard", "source": "direct"}', 1, NOW(), 'Chrome/91.0'),
    (?, '{"page": "profile", "source": "menu"}', 1, NOW(), 'Firefox/89.0')
ON DUPLICATE KEY UPDATE 
    session_data = JSON_MERGE_PATCH(session_data, VALUES(session_data)),
    page_views = page_views + VALUES(page_views),
    last_visit = GREATEST(last_visit, VALUES(last_visit)),
    browser_info = VALUES(browser_info),
    updated_at = CURRENT_TIMESTAMP;

-- Phase 6: REPLACE statement
REPLACE INTO cache_entries (cache_key, cache_value, expires_at)
VALUES 
    ('user:123:profile', '{"name": "John", "email": "john@test.com"}', DATE_ADD(NOW(), INTERVAL 1 HOUR)),
    ('user:124:profile', '{"name": "Jane", "email": "jane@test.com"}', DATE_ADD(NOW(), INTERVAL 1 HOUR));

-- Phase 7: CTEs and Window Functions
WITH RECURSIVE department_tree AS (
    -- Base case: top-level departments
    SELECT 
        id, 
        name, 
        parent_id, 
        0 as level,
        CAST(name AS CHAR(1000)) as hierarchy_path,
        JSON_ARRAY(id) as id_path
    FROM departments 
    WHERE parent_id IS NULL
    
    UNION ALL
    
    -- Recursive case: child departments
    SELECT 
        d.id,
        d.name,
        d.parent_id,
        dt.level + 1,
        CONCAT(dt.hierarchy_path, ' > ', d.name) as hierarchy_path,
        JSON_ARRAY_APPEND(dt.id_path, '$', d.id) as id_path
    FROM departments d
    INNER JOIN department_tree dt ON d.parent_id = dt.id
    WHERE dt.level < 5  -- Prevent infinite recursion
), 
employee_stats AS (
    SELECT 
        e.id,
        e.name,
        e.salary,
        e.department_id,
        dt.hierarchy_path,
        dt.level as dept_level,
        
        -- Window functions
        ROW_NUMBER() OVER (PARTITION BY e.department_id ORDER BY e.salary DESC) as salary_rank,
        DENSE_RANK() OVER (ORDER BY e.salary DESC) as company_rank,
        LAG(e.salary, 1) OVER (PARTITION BY e.department_id ORDER BY e.hire_date) as prev_salary,
        LEAD(e.salary, 1) OVER (PARTITION BY e.department_id ORDER BY e.hire_date) as next_salary,
        
        -- Frame specifications
        AVG(e.salary) OVER (PARTITION BY e.department_id) as dept_avg_salary,
        SUM(e.salary) OVER (
            PARTITION BY e.department_id 
            ORDER BY e.hire_date 
            ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
        ) as running_dept_payroll,
        
        -- JSON aggregation with window
        JSON_ARRAYAGG(e.skill) OVER (
            PARTITION BY e.department_id
            ORDER BY e.salary DESC 
            ROWS BETWEEN 2 PRECEDING AND 2 FOLLOWING
        ) as peer_skills
        
    FROM employees e
    INNER JOIN department_tree dt ON e.department_id = dt.id
)
SELECT 
    es.name,
    es.salary,
    es.hierarchy_path,
    es.salary_rank,
    es.company_rank,
    ROUND(es.dept_avg_salary, 2) as avg_dept_salary,
    es.running_dept_payroll,
    JSON_LENGTH(es.peer_skills) as peer_skills_count
FROM employee_stats es
WHERE es.salary_rank <= 3
ORDER BY es.dept_level, es.department_id, es.salary_rank
LIMIT 50;

-- Phase 8: DDL with indexes, generated columns, and table options
CREATE TABLE products (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    base_price DECIMAL(10,2) NOT NULL,
    discount_percentage DECIMAL(5,2) DEFAULT 0.00,
    tax_rate DECIMAL(5,4) DEFAULT 0.0825,
    
    -- Generated columns
    discounted_price DECIMAL(10,2) GENERATED ALWAYS AS (
        CASE 
            WHEN discount_percentage > 0 
            THEN base_price * (1 - discount_percentage / 100)
            ELSE base_price 
        END
    ) VIRTUAL,
    
    final_price DECIMAL(10,2) GENERATED ALWAYS AS (
        discounted_price * (1 + tax_rate)
    ) STORED,
    
    search_vector TEXT GENERATED ALWAYS AS (
        CONCAT_WS(' ', LOWER(name), LOWER(description))
    ) STORED,
    
    -- JSON column with generated extraction
    metadata JSON,
    category_name VARCHAR(100) GENERATED ALWAYS AS (
        JSON_UNQUOTE(metadata->'$.category')
    ) VIRTUAL,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT chk_positive_price CHECK (base_price > 0),
    CONSTRAINT chk_valid_discount CHECK (discount_percentage >= 0 AND discount_percentage <= 100),
    
    -- Index definitions
    INDEX idx_category_price (category_name, final_price),
    INDEX idx_created (created_at),
    FULLTEXT INDEX ft_search (name, description, search_vector)
) ENGINE=InnoDB 
  DEFAULT CHARSET=utf8mb4 
  COLLATE=utf8mb4_unicode_ci
  ROW_FORMAT=DYNAMIC;

-- Additional indexes with MySQL-specific options
CREATE UNIQUE INDEX uk_product_name ON products (name) USING BTREE;
CREATE INDEX idx_price_range ON products (base_price, discounted_price) USING BTREE;
CREATE SPATIAL INDEX sp_location ON venues (coordinates) USING RTREE;

-- ALTER TABLE with MySQL options
ALTER TABLE products 
ADD COLUMN weight DECIMAL(8,3),
ADD COLUMN dimensions JSON,
MODIFY COLUMN description TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
ADD CONSTRAINT fk_category 
    FOREIGN KEY (category_name) 
    REFERENCES categories(name) 
    ON DELETE SET NULL 
    ON UPDATE CASCADE,
ALGORITHM=INSTANT,
LOCK=NONE;