-- +goose Up
CREATE TABLE "comments" (
    "id" UUID PRIMARY KEY,
    "content" text NOT NULL,
    "author_id" UUID NOT NULL,
    "issue_id" UUID NOT NULL,
    "created_at" timestamp (3) NOT NULL,
    "updated_at" timestamp (3),
    CONSTRAINT "comments_author_id_fkey" FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    CONSTRAINT "comments_issue_id_fkey" FOREIGN KEY ("issue_id") REFERENCES "issues" ("id") ON DELETE CASCADE
);