-- name: CreateFlow :exec
INSERT INTO flows(product_id, name, description, created_at, updated_at)
VALUES(@product_id, @name, @description, @time_stamp, @time_stamp);

-- name: GetFlowsByProduct :many
SELECT id, name, description, created_at, updated_at FROM flows WHERE product_id = @product_id;

-- name: GetFlow :one
SELECT id, name, description, created_at, updated_at FROM flows WHERE id = @id;

-- name: UpdateFlow :exec
UPDATE flows SET name = @name, description = @description, updated_at = @updated_at WHERE id = @id RETURNING id;

-- name: DeleteFlow :exec
DELETE FROM flows WHERE id = @id RETURNING id;

-- name: GetFlowSteps :many
SELECT id, flow_id, current, next FROM flow_steps WHERE flow_id = @flow_id;

-- name: CreateFlowStep :exec
INSERT INTO flow_steps(flow_id, current, next) VALUES(@flow_id, @current, @next) RETURNING id;

-- name: DeleteFlowStep :exec
DELETE FROM flow_steps WHERE id = @id RETURNING id;