-- name: CreateActivityLog :one
INSERT INTO "activity_logs" (
    "id",
    "organisation_id",
    "project_id",
    "entity_type",
    "entity_id",
    "action",
    "actor_id",
    "metadata"
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: ListActivityLogsByOrg :many
SELECT *
FROM "activity_logs"
WHERE "organisation_id" = $1
AND ("created_at" < $2::timestamp OR $2::timestamp IS NULL)
ORDER BY "created_at" DESC
LIMIT $3;

-- name: ListActivityLogsByProject :many
SELECT *
FROM "activity_logs"
WHERE "project_id" = $1
AND ("created_at" < $2::timestamp OR $2::timestamp IS NULL)
ORDER BY "created_at" DESC
LIMIT $3;

-- name: ListActivityLogsByEntity :many
SELECT *
FROM "activity_logs"
WHERE "entity_id" = $1
AND ("created_at" < $2::timestamp OR $2::timestamp IS NULL)
ORDER BY "created_at" DESC
LIMIT $3;