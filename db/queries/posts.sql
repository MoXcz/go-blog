-- name: GetPosts :many
SELECT * FROM posts;

-- name: GetPost :one
SELECT * FROM posts
WHERE title = ?;

-- name: CreatePost :one
INSERT INTO posts (title, content)
VALUES (?, ?)
RETURNING *;
