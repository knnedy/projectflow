-- +goose Up
CREATE TABLE "organisations" (
    "id" UUID PRIMARY KEY,  
    "name" text NOT NULL,
    "created_at" timestamp (3) NOT NULL,
    "updated_at" timestamp (3)
);