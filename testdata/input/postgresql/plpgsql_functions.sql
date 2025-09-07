-- Simple DO block
DO $$ BEGIN RAISE NOTICE 'Hello from PL/pgSQL!'; END $$;

-- DO block with LANGUAGE
DO $block$ DECLARE counter INTEGER := 0; BEGIN SELECT COUNT(*) INTO counter FROM users WHERE active = true; RAISE NOTICE 'Active users: %', counter; END $block$ LANGUAGE plpgsql;

-- Basic function
CREATE FUNCTION get_user_count() RETURNS INTEGER AS $$ SELECT COUNT(*) FROM users; $$ LANGUAGE SQL;

-- Function with parameters and modifiers
CREATE OR REPLACE FUNCTION add_numbers(a INTEGER, b INTEGER) RETURNS INTEGER AS $$ SELECT a + b; $$ LANGUAGE SQL IMMUTABLE STRICT;

-- Function returning table
CREATE FUNCTION get_active_users() RETURNS TABLE(user_id INTEGER, username TEXT, email TEXT) AS $$ SELECT id, name, email FROM users WHERE active = true ORDER BY name; $$ LANGUAGE SQL STABLE;

-- Complex PL/pgSQL function with DECLARE
CREATE OR REPLACE FUNCTION update_user_stats(p_user_id INTEGER) RETURNS VOID AS $func$ DECLARE user_count INTEGER; last_login_date DATE; BEGIN SELECT COUNT(*), MAX(login_date) INTO user_count, last_login_date FROM user_sessions WHERE user_id = p_user_id; UPDATE users SET session_count = user_count, last_login = last_login_date, updated_at = NOW() WHERE id = p_user_id; IF NOT FOUND THEN RAISE EXCEPTION 'User % not found', p_user_id; END IF; END $func$ LANGUAGE plpgsql VOLATILE SECURITY DEFINER;

-- Trigger function
CREATE OR REPLACE FUNCTION audit_user_changes() RETURNS TRIGGER AS $trigger$ BEGIN IF TG_OP = 'DELETE' THEN INSERT INTO user_audit (user_id, operation, old_data, changed_at) VALUES (OLD.id, TG_OP, row_to_json(OLD), NOW()); RETURN OLD; ELSIF TG_OP = 'UPDATE' THEN INSERT INTO user_audit (user_id, operation, old_data, new_data, changed_at) VALUES (NEW.id, TG_OP, row_to_json(OLD), row_to_json(NEW), NOW()); RETURN NEW; ELSIF TG_OP = 'INSERT' THEN INSERT INTO user_audit (user_id, operation, new_data, changed_at) VALUES (NEW.id, TG_OP, row_to_json(NEW), NOW()); RETURN NEW; END IF; RETURN NULL; END $trigger$ LANGUAGE plpgsql VOLATILE SECURITY DEFINER;

-- Function with all modifiers
CREATE FUNCTION calculate_hash(input_text TEXT) RETURNS TEXT AS $$ SELECT MD5(input_text); $$ LANGUAGE SQL IMMUTABLE STRICT LEAKPROOF PARALLEL SAFE COST 1;

-- Function with SETOF and complex return
CREATE OR REPLACE FUNCTION get_user_hierarchy(root_user_id INTEGER) RETURNS SETOF users AS $hierarchy$ WITH RECURSIVE user_tree AS ( SELECT * FROM users WHERE id = root_user_id UNION ALL SELECT u.* FROM users u JOIN user_tree ut ON u.manager_id = ut.id ) SELECT * FROM user_tree ORDER BY id; $hierarchy$ LANGUAGE SQL STABLE COST 100 ROWS 50;