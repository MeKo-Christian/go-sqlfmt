-- Advanced PL/pgSQL function with exception handling
CREATE OR REPLACE FUNCTION
  process_payment(p_user_id INTEGER, p_amount DECIMAL) RETURNS VOID AS $$ DECLARE v_balance DECIMAL; v_user_name TEXT; BEGIN SELECT balance, name INTO v_balance, v_user_name FROM users WHERE id = p_user_id; IF v_balance < p_amount THEN RAISE EXCEPTION 'Insufficient balance for user %', v_user_name; END IF; UPDATE users SET balance = balance - p_amount WHERE id = p_user_id; INSERT INTO transactions (user_id, amount, type) VALUES (p_user_id, p_amount, 'payment'); EXCEPTION WHEN insufficient_privilege THEN RAISE NOTICE 'Access denied for user %', p_user_id; WHEN OTHERS THEN RAISE NOTICE 'Error processing payment: %', SQLERRM; ROLLBACK; END $$ LANGUAGE plpgsql;

-- PL/pgSQL function with LOOP and EXIT
CREATE OR REPLACE FUNCTION
  generate_series_custom(p_start INTEGER, p_end INTEGER) RETURNS SETOF INTEGER AS $$ DECLARE v_current INTEGER := p_start; BEGIN LOOP EXIT WHEN v_current > p_end; RETURN NEXT v_current; v_current := v_current + 1; END LOOP; RETURN; END $$ LANGUAGE plpgsql IMMUTABLE;

-- PL/pgSQL function with WHILE loop
CREATE OR REPLACE FUNCTION
  factorial(n INTEGER) RETURNS BIGINT AS $$ DECLARE result BIGINT := 1; counter INTEGER := n; BEGIN WHILE counter > 1 LOOP result := result * counter; counter := counter - 1; END LOOP; RETURN result; END $$ LANGUAGE plpgsql IMMUTABLE STRICT;

-- PL/pgSQL function with FOR loop
CREATE OR REPLACE FUNCTION
  sum_even_numbers(p_max INTEGER) RETURNS INTEGER AS $$ DECLARE v_sum INTEGER := 0; v_num INTEGER; BEGIN FOR v_num IN 1..p_max LOOP IF v_num % 2 = 0 THEN v_sum := v_sum + v_num; END IF; END LOOP; RETURN v_sum; END $$ LANGUAGE plpgsql IMMUTABLE;

-- PL/pgSQL function with nested IF/ELSIF/ELSE
CREATE OR REPLACE FUNCTION
  calculate_discount(p_total DECIMAL, p_customer_type TEXT) RETURNS DECIMAL AS $$ DECLARE v_discount DECIMAL := 0; BEGIN IF p_customer_type = 'premium' THEN IF p_total > 1000 THEN v_discount := 0.20; ELSIF p_total > 500 THEN v_discount := 0.15; ELSE v_discount := 0.10; END IF; ELSIF p_customer_type = 'regular' THEN IF p_total > 500 THEN v_discount := 0.05; END IF; ELSE v_discount := 0; END IF; RETURN p_total * (1 - v_discount); END $$ LANGUAGE plpgsql IMMUTABLE;

-- PL/pgSQL function with FOREACH
CREATE OR REPLACE FUNCTION
  array_sum(p_numbers INTEGER[]) RETURNS INTEGER AS $$ DECLARE v_sum INTEGER := 0; v_num INTEGER; BEGIN FOREACH v_num IN ARRAY p_numbers LOOP v_sum := v_sum + v_num; END LOOP; RETURN v_sum; END $$ LANGUAGE plpgsql IMMUTABLE;

-- Complex trigger function with multiple exception handlers
CREATE OR REPLACE FUNCTION
  validate_and_audit_user() RETURNS TRIGGER AS $$ DECLARE v_error_msg TEXT; BEGIN IF NEW.email IS NULL OR NEW.email = '' THEN RAISE EXCEPTION 'Email cannot be empty'; END IF; IF NEW.age < 18 THEN RAISE EXCEPTION 'User must be at least 18 years old'; END IF; INSERT INTO user_audit (user_id, action, timestamp) VALUES (NEW.id, TG_OP, NOW()); RETURN NEW; EXCEPTION WHEN unique_violation THEN RAISE NOTICE 'Duplicate email detected: %', NEW.email; RETURN NULL; WHEN check_violation THEN v_error_msg := SQLERRM; RAISE NOTICE 'Validation error: %', v_error_msg; RETURN NULL; WHEN OTHERS THEN RAISE WARNING 'Unexpected error in trigger: %', SQLERRM; RETURN NULL; END $$ LANGUAGE plpgsql;