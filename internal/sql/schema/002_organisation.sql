-- +goose Up
CREATE TABLE "organisation" (
    "id" UUID PRIMARY KEY,  
    "name" text NOT NULL,
    "created_at" timestamp (3) NOT NULL,
    "updated_at" timestamp (3),
    "owner_id" UUID NOT NULL,
    CONSTRAINT "organisation_owner_id_fkey" FOREIGN KEY ("owner_id") REFERENCES "user" ("id") ON DELETE CASCADE
);