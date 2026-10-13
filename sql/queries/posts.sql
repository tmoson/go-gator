-- name: CreatePost :exec
INSERT INTO posts (id, title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
ON CONFLICT (url) DO NOTHING
RETURNING *;

-- name: GetPostsForUser :many
WITH followed_feeds AS (
     SELECT * FROM follow
     WHERE follow.user_id = $1
)
SELECT posts.*
FROM posts
INNER JOIN followed_feeds ON followed_feeds.feed_id = posts.feed_id
ORDER BY posts.published_at DESC
LIMIT $2;
