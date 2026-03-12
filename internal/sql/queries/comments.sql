-- name: CreateComment :one
INSERT INTO "comments" (
    "id",
    "content",
    "author_id",
    "issue_id"
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetCommentById :one
SELECT *
FROM "comments"
WHERE "id" = $1;

-- name: ListCommentsByIssueId :many
SELECT *
FROM "comments"
WHERE "issue_id" = $1
ORDER BY "created_at" ASC;

-- name: UpdateComment :one
UPDATE "comments"
SET
    "content" = $2,
    "updated_at" = now()
WHERE "id" = $1
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM "comments"
WHERE "id" = $1;