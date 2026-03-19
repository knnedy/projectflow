-- +goose Up
CREATE TYPE member_role AS ENUM ('OWNER', 'ADMIN', 'MEMBER');

CREATE TABLE "members" (
    "id" UUID PRIMARY KEY,
    "role" member_role NOT NULL,
    "user_id" UUID NOT NULL,
    "organisation_id" UUID NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP(3),
    CONSTRAINT "members_user_id_fkey"
        FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
    CONSTRAINT "members_organisation_id_fkey"
        FOREIGN KEY ("organisation_id") REFERENCES "organisations" ("id") ON DELETE CASCADE
);

-- +goose Down
DROP TABLE "members";
DROP TYPE member_role;