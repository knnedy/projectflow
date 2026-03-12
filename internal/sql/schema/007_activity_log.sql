-- +goose Up
CREATE TABLE "activity_logs" (
    "id" UUID PRIMARY KEY,
    "action" TEXT NOT NULL,
    "user_id" UUID NOT NULL,
    "project_id" UUID NOT NULL,
    "target_id" UUID NOT NULL,
    "TIMESTAMP" TIMESTAMP (3) NOT NULL,
    CONSTRAINT "activity_log_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON DELETE CASCADE,
    CONSTRAINT "activity_log_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "project" ("id") ON DELETE CASCADE,
    CONSTRAINT "activity_log_target_id_fkey" FOREIGN KEY ("target_id") REFERENCES "issue" ("id") ON DELETE CASCADE
);