-- +goose Up
CREATE TABLE "issues" (
    "id" UUID PRIMARY KEY,
    "title" TEXT NOT NULL,
    "description" TEXT,
    "status" TEXT NOT NULL,
    "priority" TEXT NOT NULL,
    "project_id" UUID NOT NULL,
    "reporter_id" UUID NOT NULL,
    "assignee_id" UUID,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP (3),
    CONSTRAINT "issue_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "project" ("id") ON DELETE CASCADE,
    CONSTRAINT "issue_reporter_id_fkey" FOREIGN KEY ("reporter_id") REFERENCES "user" ("id") ON DELETE CASCADE,
    CONSTRAINT "issue_assignee_id_fkey" FOREIGN KEY ("assignee_id") REFERENCES "user" ("id") ON DELETE SET NULL
);