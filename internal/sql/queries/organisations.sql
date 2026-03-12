-- name: CreateOrganisation :one
INSERT INTO "organisations" (
    "id",
    "name",
    "owner_id"
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetOrganisationById :one
SELECT *
FROM "organisations"
WHERE "id" = $1;

-- name: GetOrganisationsByOwner :many
SELECT *
FROM "organisations"
WHERE "owner_id" = $1
ORDER BY "created_at" DESC;

-- name: UpdateOrganisation :one
UPDATE "organisations"
SET
    "name" = $2,
    "updated_at" = now()
WHERE "id" = $1
RETURNING *;

-- name: DeleteOrganisation :exec
DELETE FROM "organisations"
WHERE "id" = $1;