-- +goose Up
CREATE TABLE "users" (
	"id" UUID PRIMARY KEY,
	"name" text NOT NULL,
	"email" text NOT NULL,
	"password" text NOT NULL,
	"created_at" timestamp (3) NOT NULL,
	"updated_at" timestamp (3),
	CONSTRAINT "users_email_key" UNIQUE ("email")
);