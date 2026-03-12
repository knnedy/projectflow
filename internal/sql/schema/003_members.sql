-- +goose Up
CREATE TABLE "members" (
    "id" UUID PRIMARY KEY,
    "role" TEXT NOT NULL,
    "user_id" UUID NOT NULL,
    "organisation_id" UUID NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP (3),
    CONSTRAINT "member_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON DELETE CASCADE,
    CONSTRAINT "member_organisation_id_fkey" FOREIGN KEY ("organisation_id") REFERENCES "organisation" ("id") ON DELETE CASCADE
);