-- name: CreateInvitation :one
INSERT INTO "invitations" (
    "id",
    "email",
    "organisation_id",
    "role",
    "token",
    "invited_by",
    "expires_at"
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetInvitationByToken :one
SELECT * FROM "invitations"
WHERE "token" = $1
AND "accepted_at" IS NULL
AND "expires_at" > now();

-- name: GetInvitationByEmailAndOrg :one
SELECT * FROM "invitations"
WHERE "email" = $1
AND "organisation_id" = $2
AND "accepted_at" IS NULL
AND "expires_at" > now();

-- name: AcceptInvitation :one
UPDATE "invitations"
SET "accepted_at" = now()
WHERE "token" = $1
RETURNING *;

-- name: DeleteExpiredInvitations :exec
DELETE FROM "invitations"
WHERE "expires_at" < now()
AND "accepted_at" IS NULL;