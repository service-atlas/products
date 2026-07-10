-- name: CreateCapability :one
INSERT INTO capabilities(product_id, name, description, created_at, updated_at)
VALUES(@product_id, @name, @description, @timestamp, @timestamp)
RETURNING *;

-- name: GetCapability :one
SELECT id, product_id, name, description, created_at, updated_at FROM capabilities WHERE id = @id;

-- name: GetCapabilitiesByFlow :many
SELECT DISTINCT
    c.id,
    c.product_id,
    c.name,
    c.description,
    c.created_at,
    c.updated_at,
    f.name AS flow_name
FROM capabilities c
join capability_steps cs on c.id = cs.capability_id
join flow_steps fs on cs.flow_step_id = fs.id
join flows f on fs.flow_id = f.id
where f.id = @flow_id;

-- name: GetCapabilitiesByProduct :many
SELECT *
FROM capabilities c
WHERE product_id = @product_id;

-- name: UpdateCapability :execrows
UPDATE capabilities SET name = @name, description = @description, updated_at = @updated_at WHERE id = @id;

-- name: DeleteCapability :execrows
DELETE FROM capabilities WHERE id = @id;

-- name: GetFlow :one
SELECT name FROM flows WHERE id = @id;

-- name: GetProduct :one
SELECT name FROM products WHERE id = @id;
