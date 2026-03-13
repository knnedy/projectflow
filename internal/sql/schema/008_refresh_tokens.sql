-- +goose Up
CREATE TABLE "refresh_tokens" (
    "id" UUID PRIMARY KEY,
    "user_id" UUID NOT NULL,
    "token" TEXT NOT NULL,
    "expires_at" TIMESTAMP(3) NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT now(),
    "revoked_at" TIMESTAMP(3),
    CONSTRAINT "refresh_tokens_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    CONSTRAINT "refresh_tokens_token_key" UNIQUE ("token")
);

-- +goose Down
DROP TABLE "refresh_tokens";