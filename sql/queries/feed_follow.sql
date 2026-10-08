-- name: CreateFeedFollow :many
WITH inserted_feed_follow AS (
  INSERT INTO follow (id, created_at, updated_at, user_id, feed_id)
  VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
  )
  RETURNING *
) SELECT
    inserted_feed_follow.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM inserted_feed_follow
INNER JOIN feeds ON inserted_feed_follow.feed_id = feeds.id
INNER JOIN users ON inserted_feed_follow.user_id = users.id;

-- name: GetFeedFollowsForUser :many
WITH followed_feeds AS (
     SELECT * FROM follow
     WHERE follow.user_id = $1
)
SELECT
    feeds.name AS feed_name
FROM followed_feeds
INNER JOIN feeds ON followed_feeds.feed_id = feeds.id;


-- name: UserUnfollowFeed :exec
DELETE FROM follow
USING feeds
WHERE follow.user_id = $1 AND feeds.url = $2;
