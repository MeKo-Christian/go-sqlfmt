-- MySQL Mixed DDL/DML: Social media platform user interactions
-- Combined DDL and DML with transactions, triggers, and constraints

DELIMITER //

-- Start transaction for social platform setup
START TRANSACTION//

-- Create users table (assuming it doesn't exist)
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    bio TEXT,
    profile_picture_url VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_active (is_active)
)//

-- Create posts table
CREATE TABLE posts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    content TEXT NOT NULL,
    image_url VARCHAR(500),
    video_url VARCHAR(500),
    location VARCHAR(255),
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at),
    INDEX idx_public (is_public),
    FULLTEXT idx_content_search (content)
)//

-- Create follows table for user relationships
CREATE TABLE follows (
    id INT AUTO_INCREMENT PRIMARY KEY,
    follower_id INT NOT NULL,
    following_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_follow (follower_id, following_id),
    INDEX idx_follower (follower_id),
    INDEX idx_following (following_id),
    CHECK (follower_id != following_id)
)//

-- Create likes table
CREATE TABLE likes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    post_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    UNIQUE KEY unique_like (user_id, post_id),
    INDEX idx_user_id (user_id),
    INDEX idx_post_id (post_id)
)//

-- Create comments table
CREATE TABLE comments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    post_id INT NOT NULL,
    content TEXT NOT NULL,
    parent_id INT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_post_id (post_id),
    INDEX idx_parent_id (parent_id),
    FULLTEXT idx_comment_search (content)
)//

-- Create notifications table
CREATE TABLE notifications (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    type ENUM('like', 'comment', 'follow', 'mention') NOT NULL,
    actor_id INT NOT NULL,
    post_id INT NULL,
    comment_id INT NULL,
    message TEXT,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_actor_id (actor_id),
    INDEX idx_read (is_read),
    INDEX idx_created_at (created_at)
)//

-- Insert sample users
INSERT INTO users (username, email, password_hash, full_name, bio) VALUES
('johndoe', 'john@example.com', '$2b$10$dummy.hash.for.demo.purposes', 'John Doe', 'Software developer and tech enthusiast'),
('janedoe', 'jane@example.com', '$2b$10$dummy.hash.for.demo.purposes', 'Jane Doe', 'UX designer and coffee lover'),
('bobsmith', 'bob@example.com', '$2b$10$dummy.hash.for.demo.purposes', 'Bob Smith', 'Product manager at TechCorp'),
('alicejohnson', 'alice@example.com', '$2b$10$dummy.hash.for.demo.purposes', 'Alice Johnson', 'Data scientist and machine learning expert')//

-- Insert sample posts
INSERT INTO posts (user_id, content, image_url, location) VALUES
(1, 'Just shipped a new feature! The team worked really hard on this one. #programming #webdev', 'https://example.com/images/post1.jpg', 'San Francisco, CA'),
(2, 'Beautiful sunset from my office window today. Sometimes you need to take a moment to appreciate the little things. 🌅', 'https://example.com/images/post2.jpg', 'New York, NY'),
(3, 'Excited to announce our new product roadmap for Q2. We have some amazing features coming up! Stay tuned. 🚀', NULL, 'Austin, TX'),
(4, 'Working on some interesting ML models today. The results are promising! #MachineLearning #DataScience', NULL, 'Seattle, WA'),
(1, 'Coffee and code - the perfect combination for a productive morning. ☕💻', 'https://example.com/images/post3.jpg', NULL),
(2, 'Just finished reading "The Design of Everyday Things" by Don Norman. Highly recommend for anyone in UX! 📖', NULL, NULL)//

-- Create follow relationships
INSERT INTO follows (follower_id, following_id) VALUES
(1, 2), (1, 3), (1, 4), -- John follows Jane, Bob, Alice
(2, 1), (2, 4), -- Jane follows John, Alice
(3, 1), (3, 2), -- Bob follows John, Jane
(4, 1), (4, 2), (4, 3) -- Alice follows John, Jane, Bob

//

-- Add likes to posts
INSERT INTO likes (user_id, post_id) VALUES
(2, 1), (3, 1), (4, 1), -- John's post liked by Jane, Bob, Alice
(1, 2), (3, 2), (4, 2), -- Jane's post liked by John, Bob, Alice
(1, 3), (2, 3), (4, 3), -- Bob's post liked by John, Jane, Alice
(1, 4), (2, 4), (3, 4), -- Alice's post liked by John, Jane, Bob
(2, 5), (3, 5), (4, 5), -- John's coffee post
(1, 6), (3, 6), (4, 6) -- Jane's book post

//

-- Add comments to posts
INSERT INTO comments (user_id, post_id, content) VALUES
(2, 1, 'Congratulations on the launch! Can''t wait to try it out.'),
(3, 1, 'Great work team! The new feature looks amazing.'),
(4, 1, 'This is exactly what we needed. Well done!'),
(1, 2, 'Beautiful shot! Nature always finds a way to amaze us.'),
(3, 2, 'I need to visit NYC again. Miss those sunsets!'),
(4, 2, 'Perfect timing for this post. Just what I needed to see today.'),
(1, 3, 'Super excited for Q2! Any hints on what''s coming?'),
(2, 3, 'Your roadmaps are always spot on. Keep it up!'),
(1, 4, 'ML is fascinating! What kind of models are you working on?'),
(2, 4, 'Love seeing more data science content. Great work!'),
(3, 4, 'The field is evolving so fast. Keep us updated on your progress!')//

-- Create trigger for notifications on likes
CREATE TRIGGER notify_like
    AFTER INSERT ON likes
    FOR EACH ROW
BEGIN
    -- Don't notify if user likes their own post
    IF NEW.user_id != (SELECT user_id FROM posts WHERE id = NEW.post_id) THEN
        INSERT INTO notifications (user_id, type, actor_id, post_id, message)
        SELECT
            p.user_id,
            'like',
            NEW.user_id,
            NEW.post_id,
            CONCAT(u.username, ' liked your post')
        FROM posts p
        JOIN users u ON NEW.user_id = u.id
        WHERE p.id = NEW.post_id;
    END IF;
END//

-- Create trigger for notifications on comments
CREATE TRIGGER notify_comment
    AFTER INSERT ON comments
    FOR EACH ROW
BEGIN
    -- Don't notify if user comments on their own post
    IF NEW.user_id != (SELECT user_id FROM posts WHERE id = NEW.post_id) THEN
        INSERT INTO notifications (user_id, type, actor_id, post_id, comment_id, message)
        SELECT
            p.user_id,
            'comment',
            NEW.user_id,
            NEW.post_id,
            NEW.id,
            CONCAT(u.username, ' commented on your post')
        FROM posts p
        JOIN users u ON NEW.user_id = u.id
        WHERE p.id = NEW.post_id;
    END IF;
END//

-- Create trigger for notifications on follows
CREATE TRIGGER notify_follow
    AFTER INSERT ON follows
    FOR EACH ROW
BEGIN
    INSERT INTO notifications (user_id, type, actor_id, message)
    SELECT
        NEW.following_id,
        'follow',
        NEW.follower_id,
        CONCAT(u.username, ' started following you')
    FROM users u
    WHERE u.id = NEW.follower_id;
END//

-- Create view for user feed (posts from followed users)
CREATE VIEW user_feed AS
SELECT
    p.id,
    p.user_id,
    p.content,
    p.image_url,
    p.video_url,
    p.location,
    p.created_at,
    u.username,
    u.full_name,
    u.profile_picture_url,
    COUNT(DISTINCT l.id) as like_count,
    COUNT(DISTINCT c.id) as comment_count
FROM posts p
JOIN users u ON p.user_id = u.id
LEFT JOIN likes l ON p.id = l.post_id
LEFT JOIN comments c ON p.id = c.post_id
WHERE p.is_public = TRUE
GROUP BY p.id, p.user_id, p.content, p.image_url, p.video_url, p.location, p.created_at,
         u.username, u.full_name, u.profile_picture_url
ORDER BY p.created_at DESC//

-- Create view for unread notifications
CREATE VIEW unread_notifications AS
SELECT
    n.id,
    n.user_id,
    n.type,
    n.message,
    n.created_at,
    u.username as actor_username,
    u.profile_picture_url as actor_picture
FROM notifications n
JOIN users u ON n.actor_id = u.id
WHERE n.is_read = FALSE
ORDER BY n.created_at DESC//

-- Grant permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON users TO social_user//
GRANT SELECT, INSERT, UPDATE, DELETE ON posts TO social_user//
GRANT SELECT, INSERT, UPDATE, DELETE ON follows TO social_user//
GRANT SELECT, INSERT, UPDATE, DELETE ON likes TO social_user//
GRANT SELECT, INSERT, UPDATE, DELETE ON comments TO social_user//
GRANT SELECT, INSERT, UPDATE, DELETE ON notifications TO social_user//

COMMIT//

DELIMITER ;