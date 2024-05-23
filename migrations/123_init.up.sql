-- таблица с постами
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    allow_comments BOOLEAN NOT NULL DEFAULT TRUE
);

-- таблица с комментами
CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    post_id INT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author VARCHAR(255) NOT NULL,
    text TEXT NOT NULL CHECK (char_length(text) <= 2000),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    has_replies BOOLEAN NOT NULL DEFAULT FALSE
);

-- таблица со связями комментов
CREATE TABLE IF NOT EXISTS comment_hierarchy (
    parent_id INT NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    child_id INT NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    PRIMARY KEY (parent_id, child_id)
);

-- индексы(пытаемся в оптимизацию)
CREATE INDEX IF NOT EXISTS idx_posts_created_at ONgo r posts(created_at);
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at);
CREATE INDEX IF NOT EXISTS idx_comment_hierarchy_parent_id ON comment_hierarchy(parent_id);
CREATE INDEX IF NOT EXISTS idx_comment_hierarchy_child_id ON comment_hierarchy(child_id);


