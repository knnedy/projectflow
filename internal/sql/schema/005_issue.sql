-- +goose Up
CREATE TABLE "issue" (
    "id" UUID PRIMARY KEY,
    "title" text NOT NULL,
    "description" text,
    "status" text NOT NULL,
    "priority" text NOT NULL,
    "project_id" UUID NOT NULL,
    "reporter_id" UUID NOT NULL,
    "assignee_id" UUID,
    "created_at" timestamp (3) NOT NULL,
    "updated_at" timestamp (3),
    CONSTRAINT "issue_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "project" ("id") ON DELETE CASCADE,
    CONSTRAINT "issue_reporter_id_fkey" FOREIGN KEY ("reporter_id") REFERENCES "user" ("id") ON DELETE CASCADE,
    CONSTRAINT "issue_assignee_id_fkey" FOREIGN KEY ("assignee_id") REFERENCES "user" ("id") ON DELETE SET NULL
);