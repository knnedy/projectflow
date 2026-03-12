-- +goose Up
CREATE TABLE "projects" (
    "id" UUID PRIMARY KEY,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "organisation_id" UUID NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP (3),
    CONSTRAINT "project_organisation_id_fkey" FOREIGN KEY ("organisation_id") REFERENCES "organisation" ("id") ON DELETE CASCADE
)