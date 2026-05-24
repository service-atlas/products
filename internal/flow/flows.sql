-- name: CreateFlow :one
INSERT INTO flows(product_id, name, description, created_at, updated_at)
VALUES(@product_id, @name, @description, @time_stamp, @time_stamp)
RETURNING *;

-- name: GetFlowsByProduct :many
SELECT id, product_id, name, description, created_at, updated_at FROM flows WHERE product_id = @product_id;

-- name: GetFlow :one
SELECT id, product_id, name, description, created_at, updated_at FROM flows WHERE id = @id;

-- name: UpdateFlow :execrows
UPDATE flows SET name = @name, description = @description, updated_at = @updated_at WHERE id = @id;

-- name: DeleteFlow :execrows
DELETE FROM flows WHERE id = @id;

-- name: GetFlowSteps :many
SELECT id, flow_id, target, protocol, current, next, created_at, updated_at FROM flow_steps WHERE flow_id = @flow_id;

-- name: CreateFlowStep :execrows
INSERT INTO flow_steps(flow_id, current, next, target, protocol, created_at, updated_at) VALUES(@flow_id, @current, @next, @target, @protocol, @timestamp, @timestamp);

-- name: UpdateFlowStep :execrows
UPDATE flow_steps SET target = @target, protocol = @protocol, updated_at = @updated_at WHERE id = @id;

-- name: DeleteFlowStep :execrows
DELETE FROM flow_steps WHERE id = @id;

-- name: GetProductById :one
SELECT id FROM products WHERE id = @id;