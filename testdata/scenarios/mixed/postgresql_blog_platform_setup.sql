-- PostgreSQL Mixed DDL/DML: Blog platform with content management
-- Combined DDL and DML operations with transactions and constraints

-- Start transaction for blog platform setup
BEGIN;

-- Create blogs table
CREATE TABLE blogs (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    slug VARCHAR(200) UNIQUE NOT NULL,
    description TEXT,
    owner_id INTEGER NOT NULL,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
    CHECK (length(title) > 0),
    CHECK (length(slug) > 0)
);

-- Create posts table
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    blog_id INTEGER NOT NULL,
    title VARCHAR(300) NOT NULL,
    slug VARCHAR(300) NOT NULL,
    content TEXT,
    excerpt TEXT,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    published_at TIMESTAMP WITH TIME ZONE,
    author_id INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (blog_id) REFERENCES blogs(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (blog_id, slug),
    CHECK (length(title) > 0),
    CHECK (length(slug) > 0),
    CHECK (status = 'published' AND published_at IS NOT NULL OR status != 'published')
);

-- Create tags and post_tags junction table
CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    color VARCHAR(7) DEFAULT '#6B7280' CHECK (color ~ '^#[0-9A-Fa-f]{6}$'),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE post_tags (
    post_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (post_id, tag_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Create comments table
CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    post_id INTEGER NOT NULL,
    author_id INTEGER,
    author_name VARCHAR(100),
    author_email VARCHAR(255),
    content TEXT NOT NULL,
    is_approved BOOLEAN DEFAULT FALSE,
    parent_id INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE,
    CHECK (author_id IS NOT NULL OR (author_name IS NOT NULL AND author_email IS NOT NULL)),
    CHECK (length(content) > 0)
);

-- Insert sample data
INSERT INTO blogs (title, slug, description, owner_id) VALUES
('Tech Blog', 'tech-blog', 'Latest technology news and tutorials', 1),
('Personal Blog', 'personal-blog', 'My thoughts and experiences', 2);

INSERT INTO posts (blog_id, title, slug, content, excerpt, status, published_at, author_id) VALUES
(1, 'Introduction to PostgreSQL', 'introduction-postgresql',
 'PostgreSQL is a powerful open-source relational database...',
 'Learn the basics of PostgreSQL database management', 'published',
 CURRENT_TIMESTAMP - INTERVAL '2 days', 1),
(1, 'Working with JSON in PostgreSQL', 'postgresql-json',
 'PostgreSQL has excellent support for JSON data types...',
 'Explore JSON capabilities in PostgreSQL', 'published',
 CURRENT_TIMESTAMP - INTERVAL '1 day', 1),
(2, 'My Weekend Adventures', 'weekend-adventures',
 'This weekend I went hiking in the mountains...',
 'A tale of weekend adventures', 'published',
 CURRENT_TIMESTAMP - INTERVAL '6 hours', 2);

-- Insert tags
INSERT INTO tags (name, color) VALUES
('postgresql', '#336791'),
('database', '#2D5F2E'),
('tutorial', '#8B5A3C'),
('personal', '#9C27B0');

-- Associate tags with posts
INSERT INTO post_tags (post_id, tag_id) VALUES
(1, 1), (1, 2), (1, 3), -- PostgreSQL intro: postgresql, database, tutorial
(2, 1), (2, 2), (2, 3), -- JSON post: postgresql, database, tutorial
(3, 4); -- Personal post: personal

-- Add some comments
INSERT INTO comments (post_id, author_id, content, is_approved) VALUES
(1, 2, 'Great introduction! Looking forward to more PostgreSQL content.', TRUE),
(1, NULL, 'Anonymous User', 'Very helpful tutorial, thanks!', TRUE),
(2, 1, 'The JSON examples are really clear. Thanks for sharing!', TRUE);

-- Create indexes for performance
CREATE INDEX idx_posts_blog_id ON posts(blog_id);
CREATE INDEX idx_posts_author_id ON posts(author_id);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_published_at ON posts(published_at);
CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_author_id ON comments(author_id);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply triggers
CREATE TRIGGER update_blogs_updated_at
    BEFORE UPDATE ON blogs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_posts_updated_at
    BEFORE UPDATE ON posts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create view for published posts with author info
CREATE VIEW published_posts AS
SELECT
    p.id,
    p.title,
    p.slug,
    p.excerpt,
    p.content,
    p.published_at,
    p.created_at,
    b.title as blog_title,
    b.slug as blog_slug,
    u.username as author_username,
    u.email as author_email,
    ARRAY_AGG(t.name) as tags
FROM posts p
JOIN blogs b ON p.blog_id = b.id
JOIN users u ON p.author_id = u.id
LEFT JOIN post_tags pt ON p.id = pt.post_id
LEFT JOIN tags t ON pt.tag_id = t.id
WHERE p.status = 'published'
  AND b.is_public = TRUE
GROUP BY p.id, p.title, p.slug, p.excerpt, p.content, p.published_at, p.created_at,
         b.title, b.slug, u.username, u.email
ORDER BY p.published_at DESC;

-- Grant permissions
GRANT SELECT, INSERT, UPDATE ON blogs TO blog_user;
GRANT SELECT, INSERT, UPDATE ON posts TO blog_user;
GRANT SELECT, INSERT, UPDATE ON tags TO blog_user;
GRANT SELECT, INSERT, UPDATE ON post_tags TO blog_user;
GRANT SELECT, INSERT, UPDATE ON comments TO blog_user;
GRANT USAGE ON SEQUENCE blogs_id_seq TO blog_user;
GRANT USAGE ON SEQUENCE posts_id_seq TO blog_user;
GRANT USAGE ON SEQUENCE tags_id_seq TO blog_user;
GRANT USAGE ON SEQUENCE comments_id_seq TO blog_user;

COMMIT;