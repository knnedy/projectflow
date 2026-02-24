-- +goose Up
CREATE TABLE "projects" (
    "id" UUID PRIMARY KEY,
    "name" text NOT NULL,
    "description" text,
    "organisation_id" UUID NOT NULL,
    "created_at" timestamp (3) NOT NULL,
    "updated_at" timestamp (3),
    CONSTRAINT "projects_organisation_id_fkey" FOREIGN KEY ("organisation_id") REFERENCES "organisations" ("id") ON DELETE CASCADE
)