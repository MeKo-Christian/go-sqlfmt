-- SQLite DDL Essentials Test File
-- Tests Phase 7: CREATE TABLE with generated columns, CREATE INDEX, PRAGMA statements
-- Basic CREATE TABLE with generated columns (VIRTUAL)
CREATE TABLE
  products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    tax_rate DECIMAL(5, 4) DEFAULT 0.08,
    price_with_tax GENERATED ALWAYS AS (price * (1 + tax_rate)) VIRTUAL,
    slug GENERATED ALWAYS AS (LOWER(REPLACE(name, ' ', '-'))) VIRTUAL
  );

-- CREATE TABLE with STORED generated column and STRICT mode
CREATE TABLE
  inventory STRICT (
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    reorder_level INTEGER NOT NULL DEFAULT 10,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    needs_reorder GENERATED ALWAYS AS (quantity <= reorder_level) STORED,
    status_summary GENERATED ALWAYS AS (
      CASE
        WHEN quantity <= 0 THEN 'OUT_OF_STOCK'
        WHEN quantity <= reorder_level THEN 'LOW_STOCK'
        ELSE 'IN_STOCK'
      END
    ) STORED
  );

-- CREATE TABLE WITHOUT ROWID with generated columns
CREATE TABLE
  user_sessions WITHOUT ROWID (
    session_token TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    is_expired GENERATED ALWAYS AS (expires_at < unixepoch()) VIRTUAL,
    session_duration GENERATED ALWAYS AS (expires_at - created_at) STORED
  );

-- CREATE INDEX statements with IF NOT EXISTS
CREATE INDEX
  IF NOT EXISTS idx_products_name ON products(name);

CREATE UNIQUE INDEX
  IF NOT EXISTS idx_products_slug ON products(slug);

CREATE INDEX
  IF NOT EXISTS idx_inventory_status ON inventory(needs_reorder, quantity);

CREATE INDEX
  IF NOT EXISTS idx_sessions_user_active ON user_sessions(user_id)
WHERE
  NOT is_expired;

-- Complex CREATE TABLE with multiple generated columns and constraints
CREATE TABLE
  orders STRICT (
    id INTEGER PRIMARY KEY,
    customer_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL CHECK(quantity > 0),
    unit_price DECIMAL(10, 2) NOT NULL,
    discount_percent DECIMAL(5, 2) DEFAULT 0.0,
    order_date TEXT NOT NULL DEFAULT (datetime('now')),
    -- Generated columns for calculations
    subtotal GENERATED ALWAYS AS (quantity * unit_price) STORED,
    discount_amount GENERATED ALWAYS AS (subtotal * discount_percent / 100.0) VIRTUAL,
    total_amount GENERATED ALWAYS AS (subtotal - discount_amount) STORED,
    -- Generated columns for categorization
    order_size GENERATED ALWAYS AS (
      CASE
        WHEN quantity >= 100 THEN 'BULK'
        WHEN quantity >= 10 THEN 'MEDIUM'
        ELSE 'SMALL'
      END
    ) VIRTUAL,
    -- Constraints
    FOREIGN KEY (customer_id) REFERENCES customers(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
  );

-- PRAGMA statements (minimal formatting)
PRAGMA
  foreign_keys = ON;

PRAGMA
  journal_mode = WAL;

PRAGMA
  synchronous = NORMAL;

PRAGMA
  cache_size = -64000;

PRAGMA
  temp_store = MEMORY;

PRAGMA
  mmap_size = 268435456;

-- More complex PRAGMA with quoted values
PRAGMA
  table_info(orders);

PRAGMA
  index_list('products');

PRAGMA
  compile_options;

-- CREATE TABLE with JSON column and generated JSON accessors
CREATE TABLE
  user_preferences (
    user_id INTEGER PRIMARY KEY,
    settings JSON NOT NULL DEFAULT '{}',
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    -- Generated columns for JSON access
    theme GENERATED ALWAYS AS (json_extract(settings, '$.theme')) VIRTUAL,
    notifications_enabled GENERATED ALWAYS AS (
      json_extract(settings, '$.notifications.enabled')
    ) VIRTUAL,
    language GENERATED ALWAYS AS (
      COALESCE(json_extract(settings, '$.language'), 'en')
    ) STORED
  );

-- CREATE INDEX on generated columns
CREATE INDEX
  IF NOT EXISTS idx_preferences_theme ON user_preferences(theme);

CREATE INDEX
  IF NOT EXISTS idx_preferences_language ON user_preferences(language);

-- Complex table combining all features
CREATE TABLE
  audit_log STRICT (
    id INTEGER PRIMARY KEY,
    table_name TEXT NOT NULL,
    record_id INTEGER NOT NULL,
    action TEXT NOT NULL CHECK(action IN ('INSERT', 'UPDATE', 'DELETE')),
    old_values JSON,
    new_values JSON,
    changed_at REAL NOT NULL DEFAULT (julianday('now')),
    changed_by INTEGER NOT NULL,
    -- Generated columns
    change_type GENERATED ALWAYS AS (
      CASE
        WHEN old_values IS NULL THEN 'CREATE'
        WHEN new_values IS NULL THEN 'DELETE'
        ELSE 'MODIFY'
      END
    ) VIRTUAL,
    has_sensitive_data GENERATED ALWAYS AS (
      json_extract(new_values, '$.password') IS NOT NULL
      OR json_extract(new_values, '$.ssn') IS NOT NULL
      OR json_extract(old_values, '$.password') IS NOT NULL
      OR json_extract(old_values, '$.ssn') IS NOT NULL
    ) STORED,
    change_summary GENERATED ALWAYS AS (
      table_name || '.' || record_id || ' ' || change_type || ' by user ' || changed_by
    ) VIRTUAL
  );

-- Final CREATE INDEX statements
CREATE INDEX
  IF NOT EXISTS idx_audit_table_record ON audit_log(table_name, record_id);

CREATE INDEX
  IF NOT EXISTS idx_audit_changed_at ON audit_log(changed_at);

CREATE INDEX
  IF NOT EXISTS idx_audit_sensitive ON audit_log(has_sensitive_data)
WHERE
  has_sensitive_data;