-- Advanced SQLite trigger with multiple statements in BEGIN/END block
CREATE TRIGGER log_user_changes AFTER UPDATE ON users FOR EACH ROW BEGIN INSERT INTO user_audit (user_id, old_email, new_email, changed_at) VALUES (NEW.id, OLD.email, NEW.email, datetime('now')); UPDATE user_stats SET update_count = update_count + 1 WHERE user_id = NEW.id; INSERT INTO change_log (table_name, record_id, action) VALUES ('users', NEW.id, 'UPDATE'); END;

-- BEFORE INSERT trigger with validation logic
CREATE TRIGGER validate_product_before_insert BEFORE INSERT ON products FOR EACH ROW BEGIN SELECT CASE WHEN NEW.price <= 0 THEN RAISE(ABORT, 'Price must be positive') WHEN NEW.stock < 0 THEN RAISE(ABORT, 'Stock cannot be negative') WHEN LENGTH(NEW.name) < 3 THEN RAISE(ABORT, 'Product name too short') END; INSERT INTO product_audit (product_id, action, timestamp) VALUES (NEW.id, 'INSERT', datetime('now')); END;

-- INSTEAD OF trigger for view
CREATE TRIGGER update_user_view INSTEAD OF UPDATE ON user_summary_view FOR EACH ROW BEGIN UPDATE users SET name = NEW.name, email = NEW.email, updated_at = datetime('now') WHERE id = NEW.id; UPDATE user_profiles SET bio = NEW.bio WHERE user_id = NEW.id; INSERT INTO view_update_log (view_name, user_id, timestamp) VALUES ('user_summary_view', NEW.id, datetime('now')); END;

-- AFTER DELETE trigger with cascading actions
CREATE TRIGGER cleanup_user_data AFTER DELETE ON users FOR EACH ROW BEGIN DELETE FROM user_sessions WHERE user_id = OLD.id; DELETE FROM user_preferences WHERE user_id = OLD.id; DELETE FROM user_notifications WHERE user_id = OLD.id; INSERT INTO deleted_users (id, email, deleted_at) VALUES (OLD.id, OLD.email, datetime('now')); UPDATE statistics SET total_users = total_users - 1; END;

-- Complex trigger with conditional logic
CREATE TRIGGER manage_order_status AFTER UPDATE OF status ON orders FOR EACH ROW BEGIN SELECT CASE WHEN NEW.status = 'completed' AND OLD.status != 'completed' THEN (UPDATE customers SET total_orders = total_orders + 1 WHERE id = NEW.customer_id) WHEN NEW.status = 'cancelled' AND OLD.status != 'cancelled' THEN (INSERT INTO cancelled_orders (order_id, reason, cancelled_at) VALUES (NEW.id, 'User cancelled', datetime('now'))) END; INSERT INTO order_status_history (order_id, old_status, new_status, changed_at) VALUES (NEW.id, OLD.status, NEW.status, datetime('now')); END;

-- BEFORE UPDATE trigger with validation and auto-update
CREATE TRIGGER update_product_timestamp BEFORE UPDATE ON products FOR EACH ROW BEGIN SELECT CASE WHEN NEW.price < OLD.price * 0.5 THEN RAISE(ABORT, 'Price reduction exceeds 50%') WHEN NEW.stock > OLD.stock + 1000 THEN RAISE(ABORT, 'Stock increase too large') END; SELECT NEW.id AS id, NEW.name AS name, NEW.price AS price, NEW.stock AS stock, datetime('now') AS updated_at; END;
