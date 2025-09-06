select id,name,email from users where active=true order by name;

SELECT u.id,u.username, u.email,p.first_name,p.last_name,p.created_at FROM users u LEFT JOIN profiles p ON u.id = p.user_id WHERE u.active = true AND u.email_verified = true AND p.created_at > '2023-01-01' ORDER BY u.username ASC, p.created_at DESC;