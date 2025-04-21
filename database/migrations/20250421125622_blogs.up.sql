CREATE TABLE IF NOT EXISTS blogs (
                                     id VARCHAR(50) PRIMARY KEY,
    image_url TEXT,
    title VARCHAR(256) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    author VARCHAR(50) NOT NULL,
    body TEXT,
    blog_category VARCHAR(50) NOT NULL,
    FOREIGN KEY (blog_category) REFERENCES blog_categories(id)
    );