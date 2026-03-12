-- name: CreateOrganization :one
INSERT INTO "organizations" (
    "id",
    "name",
    "owner_id"
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetOrganizationById :one
SELECT *
FROM "organizations"
WHERE "id" = $1;

-- name: GetOrganizationsByOwner :many
SELECT *
FROM "organizations"
WHERE "owner_id" = $1
ORDER BY "created_at" DESC;

-- name: UpdateOrganization :one
UPDATE "organizations"
SET
    "name" = $2,
    "updated_at" = now()
WHERE "id" = $1
RETURNING *;

-- name: DeleteOrganization :exec
DELETE FROM "organizations"
WHERE "id" = $1;