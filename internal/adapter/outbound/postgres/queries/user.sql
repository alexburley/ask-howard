-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: FindUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: FindUserByID :one
SELECT * FROM users
WHERE id = $1;
