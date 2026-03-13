-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
VALUES ($1, $2, $3, $4, now())
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1
AND revoked_at IS NULL
AND expires_at > now();

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE token = $1;

-- name: RevokeAllUserTokens :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE user_id = $1
AND revoked_at IS NULL;