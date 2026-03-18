-- name: CreateUser :one

INSERT INTO users (name,email,password_hash,token_version,verified_expiry ) values ($1, $2, $3, $4, $5) returning *;

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
SELECT id, name, email, password_hash, token_version, is_verified FROM users WHERE email= $1;


-- name: ConsumeRefreshTokenByID :one
DELETE FROM refresh_tokens WHERE token_id= $1 returning user_id, hashed_token, expires_at, created_at, token_id;


-- name: DeleteAllRefreshTokenByUserID :exec
DELETE FROM refresh_tokens WHERE user_id= $1;

-- name: GetRefreshTokenByID :one
SELECT * FROM refresh_tokens WHERE token_id= $1;


-- name: GetUserByID :one
SELECT id, name, email, password_hash, token_version, is_verified FROM users WHERE id= $1;

-- name: GetVerificationByToken :one
SELECT id, user_id, selector, expires_at, verifier_hash FROM verification_tokens WHERE selector= $1;

-- name: ConsumeVerification :one
DELETE FROM verification_tokens WHERE id= $1 returning id, user_id, selector, created_at, expires_at;

-- name: SaveOneTimeToken :one
INSERT INTO verification_tokens(user_id, selector, expires_at, verifier_hash) values ($1, $2, $3, $4) returning id, user_id, selector, expires_at, verifier_hash;

-- name: UpdateVerifiedState :exec
UPDATE users SET is_verified= true WHERE id= $1;

-- name: UpdateVerificationUsers :exec
UPDATE users SET verified_expiry= $1 WHERE id= $2;

-- name: SaveResetPassword :one
INSERT INTO password_reset(verifier_hash, user_id, expiry, selector) values ($1, $2, $3, $4) returning user_id, verifier_hash, expiry, selector ;

-- name: GetResetPasswordBySelector :one
SELECT verifier_hash, user_id, is_used,expiry FROM password_reset WHERE selector=$1;

-- name: UpdateResetPasswordStatus :exec
UPDATE password_reset SET is_used= true WHERE selector= $1;

-- name: UpdatePassword :one
UPDATE users SET password_hash= $1 WHERE id= $2 returning name; 

-- name: ConsumePasswordReset :one
UPDATE password_reset SET is_used= true WHERE selector= $1 AND is_used= false AND expiry > now() returning user_id, verifier_hash;