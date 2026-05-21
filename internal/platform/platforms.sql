
-- name: GetPlatforms :many
SELECT id, name, description, created_at, updated_at FROM platforms order by id;

-- name: GetPlatform :one
SELECT id, name, description, created_at, updated_at FROM platforms WHERE id = @id;

-- name: CreatePlatform :one
INSERT INTO platforms (name, description, created_at, updated_at)
VALUES (@name, @description, @timeStamp, @timeStamp)
RETURNING *;;

-- name: UpdatePlatform :one
UPDATE platforms SET name = @name, description = @description, updated_at = @updated_at WHERE id = @id RETURNING id;

-- name: DeletePlatform :one
DELETE FROM platforms WHERE id = @id RETURNING id;