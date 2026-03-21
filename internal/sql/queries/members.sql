-- name: CreateMember :one
INSERT INTO "members" (
    "id",
    "role",
    "user_id",
    "organisation_id"
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetMemberById :one
SELECT *
FROM "members"
WHERE "id" = $1;

-- name: GetMemberByUserAndOrg :one
SELECT *
FROM "members"
WHERE "user_id" = $1
  AND "organisation_id" = $2;

-- name: GetMembersByOrg :many
SELECT *
FROM "members"
WHERE "organisation_id" = $1
ORDER BY "created_at" ASC;

-- name: GetOwnerAndAdminCountByOrg :one
SELECT COUNT(*) 
FROM "members"
WHERE "organisation_id" = $1
AND "role" IN ('OWNER', 'ADMIN');

-- name: UpdateMember :one
UPDATE "members"
SET
    "role" = $2,
    "updated_at" = now()
WHERE "id" = $1
RETURNING *;

-- name: DeleteMember :exec
DELETE FROM "members"
WHERE "id" = $1;