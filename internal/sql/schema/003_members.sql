-- +goose Up
CREATE TABLE "members" (
    "id" UUID PRIMARY KEY,
    "role" text NOT NULL,
    "user_id" UUID NOT NULL,
    "organisation_id" UUID NOT NULL,
    "created_at" timestamp (3) NOT NULL,
    "updated_at" timestamp (3),
    CONSTRAINT "members_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    CONSTRAINT "members_organisation_id_fkey" FOREIGN KEY ("organisation_id") REFERENCES "organisations" ("id") ON DELETE CASCADE
);