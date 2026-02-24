-- +goose Up
CREATE TABLE "project" (
    "id" UUID PRIMARY KEY,
    "name" text NOT NULL,
    "description" text,
    "organisation_id" UUID NOT NULL,
    "created_at" timestamp (3) NOT NULL,
    "updated_at" timestamp (3),
    CONSTRAINT "project_organisation_id_fkey" FOREIGN KEY ("organisation_id") REFERENCES "organisation" ("id") ON DELETE CASCADE
)