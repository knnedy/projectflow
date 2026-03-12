-- name: CreateProject :one
INSERT INTO "projects" (
    "id",
    "name",
    "description",
    "organisation_id"
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetProjectById :one
SELECT *
FROM "projects"
WHERE "id" = $1;

-- name: GetProjectsByOrganisation :many
SELECT *
FROM "projects"
WHERE "organisation_id" = $1
ORDER BY "created_at" DESC;

-- name: UpdateProject :one
UPDATE "projects"
SET
    "name" = $2,
    "description" = $3,
    "updated_at" = now()
WHERE "id" = $1
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM "projects"
WHERE "id" = $1;