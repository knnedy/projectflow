-- +goose Up
CREATE TABLE "invitations" (
    "id" UUID PRIMARY KEY,
    "email" TEXT NOT NULL,
    "organisation_id" UUID NOT NULL,
    "role" member_role NOT NULL DEFAULT 'MEMBER',
    "token" TEXT NOT NULL,
    "invited_by" UUID NOT NULL,
    "expires_at" TIMESTAMP(3) NOT NULL,
    "accepted_at" TIMESTAMP(3),
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT now(),
    CONSTRAINT "invitations_organisation_id_fkey"
        FOREIGN KEY ("organisation_id") REFERENCES "organisations" ("id") ON DELETE CASCADE,
    CONSTRAINT "invitations_invited_by_fkey"
        FOREIGN KEY ("invited_by") REFERENCES "users" ("id") ON DELETE CASCADE,
    CONSTRAINT "invitations_token_key" UNIQUE ("token"),
    CONSTRAINT "invitations_email_org_key" UNIQUE ("email", "organisation_id")
);

-- +goose Down
DROP TABLE "invitations";