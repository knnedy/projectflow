-- +goose Up
CREATE TABLE "organisations" (
    "id" UUID PRIMARY KEY,  
    "name" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP (3),
    "owner_id" UUID NOT NULL,
    CONSTRAINT "organisation_owner_id_fkey" FOREIGN KEY ("owner_id") REFERENCES "user" ("id") ON DELETE CASCADE
);