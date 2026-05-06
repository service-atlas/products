-- name: GetProductsByPlatform :many
SELECT id, platform_id, name, description, created_at, updated_at FROM products WHERE platform_id = @platform_id;

-- name: GetProductById :one
SELECT id, platform_id, name, description, created_at, updated_at FROM products WHERE id = @id;

-- name: CreateProduct :exec
INSERT INTO products (platform_id, name, description, created_at, updated_at) VALUES (@platform_id, @name, @description, @timestamp, @timestamp);

-- name: UpdateProduct :one
UPDATE products SET platform_id = @platform_id, name = @name, description = @description, updated_at = @updated_at WHERE id = @id RETURNING id;

-- name: DeleteProduct :one
DELETE FROM products WHERE id = @id RETURNING id;