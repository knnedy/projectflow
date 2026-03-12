-- +goose Up
CREATE TABLE "projects" (
    "id" UUID PRIMARY KEY,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "organisation_id" UUID NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP(3),
    CONSTRAINT "projects_organisation_id_fkey"
        FOREIGN KEY ("organisation_id") REFERENCES "organisations" ("id") ON DELETE CASCADE
);


-- +goose Down
DROP TABLE "projects";