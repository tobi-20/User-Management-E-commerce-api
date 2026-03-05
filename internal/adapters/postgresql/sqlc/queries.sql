-- name: CreateUser :one

INSERT INTO users (name,email,password_hash,token_version) values ($1, $2, $3, $4) returning *;

-- name: CreateBrand :one

INSERT INTO brands (name) values($1) returning *;

-- name: CreateCategory :one

INSERT INTO categories (name) values($1) returning *;

-- name: CreateProduct :one

INSERT INTO products(name, description) values ($1, $2) returning *;

-- name: CreateProductVariant :one

INSERT INTO product_variants(weight, unit, price_in_kobo, stock) values ($1, $2, $3, $4) returning *;

-- name: CreateOrder :one

INSERT INTO orders(shipping_cost_kobo, raw_order_price_in_kobo, discount_type, discount_value, order_total) values($1, $2, $3, $4, $5) returning *;

-- name: CreateOrderItem :one

INSERT INTO order_items(quantity, price_in_kobo,discount_type, discount_value, item_total) values ($1, $2, $3, $4, $5) returning *;

-- name: CreateShippingRules :one

INSERT INTO shipping_rules(max_price_in_kobo, min_price_in_kobo,type, value) values ($1, $2, $3, $4) returning *;

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
