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

-- name: GetOrganisationsByUser :many
SELECT o.*
FROM "organisations" o
INNER JOIN "members" m ON m.organisation_id = o.id
WHERE m.user_id = $1
ORDER BY o.created_at DESC;

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