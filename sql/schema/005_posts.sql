-- +goose Up
CREATE TABLE posts(
  id UUID PRIMARY KEY,
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now(),
  title TEXT,
  url TEXT UNIQUE NOT NULL,
  description TEXT,
  published_at TIMESTAMP NOT NULL,
  feed_id UUID NOT NULL,
  CONSTRAINT fk_feed_id
  FOREIGN KEY (feed_id)
  REFERENCES feeds(id)
  ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;
