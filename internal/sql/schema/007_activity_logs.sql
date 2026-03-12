-- +goose Up
CREATE TABLE "activity_logs" (
    "id" UUID PRIMARY KEY,
    "action" TEXT NOT NULL,
    "user_id" UUID NOT NULL,
    "project_id" UUID NOT NULL,
    "target_id" UUID NOT NULL,
    "timestamp" TIMESTAMP(3) NOT NULL,
    CONSTRAINT "activity_logs_user_id_fkey"
        FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    CONSTRAINT "activity_logs_project_id_fkey"
        FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON DELETE CASCADE,
    CONSTRAINT "activity_logs_target_id_fkey"
        FOREIGN KEY ("target_id") REFERENCES "issues" ("id") ON DELETE CASCADE
);

-- +goose Down
DROP TABLE "activity_logs";