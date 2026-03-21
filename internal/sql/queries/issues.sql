-- name: CreateIssue :one
INSERT INTO "issues" (
    "id",
    "title",
    "description",
    "status",
    "priority",
    "project_id",
    "reporter_id",
    "assignee_id"
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetIssueById :one
SELECT *
FROM "issues"
WHERE "id" = $1;

-- name: ListIssuesByProject :many
SELECT *
FROM "issues"
WHERE "project_id" = $1
ORDER BY "created_at" DESC;

-- name: UpdateIssueStatus :one
UPDATE "issues"
SET
    "status" = $2,
    "updated_at" = now()
WHERE "id" = $1
RETURNING *;

-- name: UpdateIssueDetails :one
UPDATE "issues"
SET
    "title" = $2,
    "description" = $3,
    "priority" = $4,
    "assignee_id" = $5,
    "updated_at" = now()
WHERE "id" = $1
RETURNING *;

-- name: DeleteIssue :exec
DELETE FROM "issues"
WHERE "id" = $1;