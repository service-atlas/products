-- name: CreateCapability :one
INSERT INTO capabilities(flow_id, name, description, created_at, updated_at)
VALUES(@flow_id, @name, @description, @timestamp, @timestamp)
RETURNING *;

-- name: GetCapability :one
SELECT id, flow_id, name, description, created_at, updated_at FROM capabilities WHERE id = @id;

-- name: GetCapabilitiesByFlow :many
SELECT id, flow_id, name, description, created_at, updated_at FROM capabilities WHERE flow_id = @flow_id;

-- name: UpdateCapability :execrows
UPDATE capabilities SET name = @name, description = @description, updated_at = @updated_at WHERE id = @id;

-- name: DeleteCapability :execrows
DELETE FROM capabilities WHERE id = @id;

-- name: GetFlow :one
SELECT name FROM flows WHERE id = @id;