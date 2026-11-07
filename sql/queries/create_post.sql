-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPosts :many
SELECT posts.title, posts.url, COALESCE(posts.description, '') AS content, feeds.name AS feed_name
FROM posts
JOIN feeds ON posts.feed_id = feeds.id
ORDER BY posts.published_at DESC;

-- name: GetPostForUser :one
SELECT posts.id, posts.created_at, posts.updated_at, posts.title, posts.url, posts.description, posts.published_at, feeds.name AS feed_name, users.name AS user_name
FROM posts
JOIN feeds ON posts.feed_id = feeds.id
JOIN users ON feeds.user_id = users.id
WHERE posts.id = $1 AND users.id = $2;
