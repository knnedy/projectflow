-- name: CreateActivityLog :one
INSERT INTO "activity_logs" (
    "id",
    "action",
    "user_id",
    "project_id",
    "target_id",
    "timestamp"
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: ListActivityLogsByProjectId :many
SELECT *
FROM "activity_logs"
WHERE "project_id" = $1
ORDER BY "timestamp" DESC;