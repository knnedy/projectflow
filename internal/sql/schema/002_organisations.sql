-- +goose Up
CREATE TABLE "organisations" (
    "id" UUID PRIMARY KEY,
    "name" TEXT NOT NULL,
    "owner_id" UUID NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP(3),
    CONSTRAINT "organisations_owner_id_fkey"
        FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

-- +goose Down
DROP TABLE "organisations";