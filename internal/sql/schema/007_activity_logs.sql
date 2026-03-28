-- +goose Up
CREATE TYPE activity_entity_type AS ENUM ('ISSUE', 'PROJECT', 'MEMBER', 'ORGANISATION');
CREATE TYPE activity_action AS ENUM (
    'CREATED',
    'UPDATED',
    'DELETED',
    'STATUS_CHANGED',
    'PRIORITY_CHANGED',
    'ASSIGNED',
    'UNASSIGNED',
    'ROLE_CHANGED',
    'MEMBER_JOINED',
    'MEMBER_LEFT',
    'MEMBER_REMOVED'
);

CREATE TABLE "activity_logs" (
    "id" UUID PRIMARY KEY,
    "organisation_id" UUID NOT NULL,
    "project_id" UUID,
    "entity_type" activity_entity_type NOT NULL,
    "entity_id" UUID NOT NULL,
    "action" activity_action NOT NULL,
    "actor_id" UUID NOT NULL,
    "metadata" JSONB,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT now(),
    CONSTRAINT "activity_logs_organisation_id_fkey"
        FOREIGN KEY ("organisation_id") REFERENCES "organisations" ("id") ON DELETE CASCADE,
    CONSTRAINT "activity_logs_project_id_fkey"
        FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON DELETE CASCADE,
    CONSTRAINT "activity_logs_actor_id_fkey"
        FOREIGN KEY ("actor_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

-- +goose Down
DROP TABLE "activity_logs";
DROP TYPE activity_action;
DROP TYPE activity_entity_type;