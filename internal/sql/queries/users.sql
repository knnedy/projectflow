-- name: CreateUser :one
INSERT INTO "users" (
    "id", 
    "name", 
    "email", 
    "password"
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetUserById :one
SELECT * FROM "users" WHERE "id" = $1;

-- name: GetUserByEmail :one
SELECT * FROM "users" WHERE "email" = $1;

-- name: UpdateUserProfile :one
UPDATE "users" 
SET
    "name" = $2,
    "email" = $3,
    "password" = $4,
    "updated_at" = now()    
WHERE "id" = $1
RETURNING *;    

-- name: UpdateUserPassword :one
UPDATE "users"
SET
    "password" = $2,
    "updated_at" = now()
WHERE "id" = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM "users" WHERE "id" = $1;