-- +goose Up
CREATE TABLE "comments" (
    "id" UUID PRIMARY KEY,
    "content" TEXT NOT NULL,
    "author_id" UUID NOT NULL,
    "issue_id" UUID NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP (3),
    CONSTRAINT "comment_author_id_fkey" FOREIGN KEY ("author_id") REFERENCES "user" ("id") ON DELETE CASCADE,
    CONSTRAINT "comment_issue_id_fkey" FOREIGN KEY ("issue_id") REFERENCES "issue" ("id") ON DELETE CASCADE
);