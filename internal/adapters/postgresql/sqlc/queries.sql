-- name: CreateUser :one

INSERT INTO users (name,email,password_hash,token_version) values ($1, $2, $3, $4) returning *;

-- name: SaveRefreshToken :one
INSERT INTO refresh_tokens(user_id, hashed_token, expires_at, created_at, token_id) values ($1, $2, $3, $4, $5) returning *;

-- name: GetUserByEmail :one
SELECT id, name, email, password_hash, token_version FROM users WHERE email= $1;


-- name: DeleteRefreshTokenByID :exec
DELETE FROM refresh_tokens WHERE token_id= $1;

-- name: GetRefreshTokenByID :one
SELECT * FROM refresh_tokens WHERE token_id= $1;


-- name: GetUserByID :one
SELECT id, name, email, password_hash, token_version FROM users WHERE id= $1;
