-- PostgreSQL Stored Procedures: User management and authentication
-- Production-style procedures with error handling and logging

-- Create audit log table for tracking changes
CREATE TABLE IF NOT EXISTS audit_log (
    id SERIAL PRIMARY KEY,
    table_name TEXT NOT NULL,
    operation TEXT NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
    old_values JSONB,
    new_values JSONB,
    user_id INTEGER,
    session_id INTEGER,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Function to log audit events
CREATE OR REPLACE FUNCTION log_audit_event(
    p_table_name TEXT,
    p_operation TEXT,
    p_old_values JSONB DEFAULT NULL,
    p_new_values JSONB DEFAULT NULL,
    p_user_id INTEGER DEFAULT NULL,
    p_session_id INTEGER DEFAULT NULL,
    p_ip_address INET DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL
)
RETURNS VOID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
    INSERT INTO audit_log (
        table_name, operation, old_values, new_values,
        user_id, session_id, ip_address, user_agent
    ) VALUES (
        p_table_name, p_operation, p_old_values, p_new_values,
        p_user_id, p_session_id, p_ip_address, p_user_agent
    );
EXCEPTION
    WHEN OTHERS THEN
        -- Log audit failures to PostgreSQL log but don't fail the transaction
        RAISE WARNING 'Failed to log audit event: %', SQLERRM;
END;
$$;

-- Procedure to create a new user with validation
CREATE OR REPLACE PROCEDURE create_user(
    p_username TEXT,
    p_email TEXT,
    p_password_hash TEXT,
    p_salt TEXT,
    p_role TEXT DEFAULT 'user',
    p_created_by INTEGER DEFAULT NULL
)
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_user_id INTEGER;
    v_existing_count INTEGER;
BEGIN
    -- Input validation
    IF p_username IS NULL OR length(trim(p_username)) = 0 THEN
        RAISE EXCEPTION 'Username cannot be empty';
    END IF;

    IF p_email IS NULL OR NOT (p_email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$') THEN
        RAISE EXCEPTION 'Invalid email format';
    END IF;

    IF p_password_hash IS NULL OR length(p_password_hash) < 8 THEN
        RAISE EXCEPTION 'Password hash must be at least 8 characters';
    END IF;

    -- Check for existing username or email
    SELECT COUNT(*) INTO v_existing_count
    FROM users
    WHERE username = trim(lower(p_username)) OR email = trim(lower(p_email));

    IF v_existing_count > 0 THEN
        RAISE EXCEPTION 'Username or email already exists';
    END IF;

    -- Insert new user
    INSERT INTO users (
        username, email, password_hash, salt, role, email_verified
    ) VALUES (
        trim(lower(p_username)),
        trim(lower(p_email)),
        p_password_hash,
        p_salt,
        p_role,
        FALSE
    ) RETURNING id INTO v_user_id;

    -- Log the creation
    PERFORM log_audit_event(
        'users',
        'INSERT',
        NULL,
        jsonb_build_object(
            'id', v_user_id,
            'username', trim(lower(p_username)),
            'email', trim(lower(p_email)),
            'role', p_role
        ),
        p_created_by
    );

    -- Send welcome email (placeholder - would integrate with email service)
    RAISE NOTICE 'Welcome email sent to % for user %', p_email, v_user_id;

EXCEPTION
    WHEN unique_violation THEN
        RAISE EXCEPTION 'Username or email already exists';
    WHEN OTHERS THEN
        RAISE EXCEPTION 'Failed to create user: %', SQLERRM;
END;
$$;

-- Function to authenticate user
CREATE OR REPLACE FUNCTION authenticate_user(
    p_username_or_email TEXT,
    p_password_hash TEXT,
    p_ip_address INET DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL
)
RETURNS TABLE (
    user_id INTEGER,
    username TEXT,
    email TEXT,
    role TEXT,
    session_token TEXT
)
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_user_record RECORD;
    v_session_token TEXT;
BEGIN
    -- Find user by username or email
    SELECT id, username, email, password_hash, salt, role, is_active, email_verified
    INTO v_user_record
    FROM users
    WHERE (username = trim(lower(p_username_or_email)) OR
           email = trim(lower(p_username_or_email)))
      AND is_active = TRUE;

    -- Check if user exists
    IF NOT FOUND THEN
        -- Log failed attempt
        PERFORM log_audit_event(
            'users',
            'LOGIN_FAILED',
            NULL,
            jsonb_build_object('identifier', p_username_or_email),
            NULL, NULL, p_ip_address, p_user_agent
        );
        RETURN;
    END IF;

    -- Verify password (simplified - would use proper hashing)
    IF v_user_record.password_hash != p_password_hash THEN
        -- Log failed attempt
        PERFORM log_audit_event(
            'users',
            'LOGIN_FAILED',
            NULL,
            jsonb_build_object('user_id', v_user_record.id, 'reason', 'invalid_password'),
            v_user_record.id, NULL, p_ip_address, p_user_agent
        );
        RETURN;
    END IF;

    -- Generate session token
    v_session_token := encode(gen_random_bytes(32), 'hex');

    -- Create session
    INSERT INTO user_sessions (
        user_id, session_token, ip_address, user_agent,
        expires_at
    ) VALUES (
        v_user_record.id,
        v_session_token,
        p_ip_address,
        p_user_agent,
        CURRENT_TIMESTAMP + INTERVAL '24 hours'
    );

    -- Update last login
    UPDATE users
    SET last_login_at = CURRENT_TIMESTAMP
    WHERE id = v_user_record.id;

    -- Log successful login
    PERFORM log_audit_event(
        'users',
        'LOGIN_SUCCESS',
        NULL,
        jsonb_build_object('user_id', v_user_record.id),
        v_user_record.id, NULL, p_ip_address, p_user_agent
    );

    -- Return user info and session token
    RETURN QUERY
    SELECT
        v_user_record.id,
        v_user_record.username,
        v_user_record.email,
        v_user_record.role,
        v_session_token;

EXCEPTION
    WHEN OTHERS THEN
        RAISE EXCEPTION 'Authentication failed: %', SQLERRM;
END;
$$;

-- Procedure to update user profile
CREATE OR REPLACE PROCEDURE update_user_profile(
    p_user_id INTEGER,
    p_email TEXT DEFAULT NULL,
    p_role TEXT DEFAULT NULL,
    p_is_active BOOLEAN DEFAULT NULL,
    p_updated_by INTEGER
)
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_old_values JSONB;
    v_new_values JSONB;
BEGIN
    -- Get current values for audit
    SELECT jsonb_build_object(
        'email', email,
        'role', role,
        'is_active', is_active
    ) INTO v_old_values
    FROM users
    WHERE id = p_user_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'User not found';
    END IF;

    -- Build new values object
    v_new_values := jsonb_build_object();

    -- Update email if provided
    IF p_email IS NOT NULL THEN
        IF NOT (p_email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$') THEN
            RAISE EXCEPTION 'Invalid email format';
        END IF;
        v_new_values := v_new_values || jsonb_build_object('email', trim(lower(p_email)));
        UPDATE users SET email = trim(lower(p_email)) WHERE id = p_user_id;
    END IF;

    -- Update role if provided
    IF p_role IS NOT NULL THEN
        IF p_role NOT IN ('admin', 'moderator', 'user') THEN
            RAISE EXCEPTION 'Invalid role';
        END IF;
        v_new_values := v_new_values || jsonb_build_object('role', p_role);
        UPDATE users SET role = p_role WHERE id = p_user_id;
    END IF;

    -- Update active status if provided
    IF p_is_active IS NOT NULL THEN
        v_new_values := v_new_values || jsonb_build_object('is_active', p_is_active);
        UPDATE users SET is_active = p_is_active WHERE id = p_user_id;
    END IF;

    -- Log the update
    PERFORM log_audit_event(
        'users',
        'UPDATE',
        v_old_values,
        v_new_values,
        p_updated_by
    );

EXCEPTION
    WHEN OTHERS THEN
        RAISE EXCEPTION 'Failed to update user profile: %', SQLERRM;
END;
$$;

-- Function to get user statistics
CREATE OR REPLACE FUNCTION get_user_statistics(
    p_start_date DATE DEFAULT CURRENT_DATE - INTERVAL '30 days',
    p_end_date DATE DEFAULT CURRENT_DATE
)
RETURNS TABLE (
    total_users BIGINT,
    active_users BIGINT,
    new_users BIGINT,
    users_by_role JSONB
)
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
    RETURN QUERY
    SELECT
        (SELECT COUNT(*) FROM users) as total_users,
        (SELECT COUNT(*) FROM users WHERE is_active = TRUE) as active_users,
        (SELECT COUNT(*) FROM users WHERE created_at >= p_start_date AND created_at <= p_end_date) as new_users,
        (SELECT jsonb_object_agg(role, count)
         FROM (SELECT role, COUNT(*) as count FROM users GROUP BY role) r) as users_by_role;
END;
$$;