
-- name: GetPlatforms :many
SELECT id, name, description, created_at, updated_at FROM platforms;

-- name: GetPlatform :one
SELECT id, name, description, created_at, updated_at FROM platforms WHERE id = @id;

-- name: CreatePlatform :exec
INSERT INTO platforms (name, description, created_at, updated_at) VALUES (@name, @description, @timeStamp, @timeStamp);

-- name: UpdatePlatform :exec
UPDATE platforms SET name = @name, description = @description, updated_at = @updatedAt WHERE id = @id;

-- name: DeletePlatform :exec
DELETE FROM platforms WHERE id = @id;