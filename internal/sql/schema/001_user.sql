-- +goose Up
CREATE TABLE "user" (
	"id" UUID PRIMARY KEY,
	"name" text NOT NULL,
	"email" text NOT NULL,
	"password" text NOT NULL,
	"created_at" timestamp (3) NOT NULL,
	"updated_at" timestamp (3),
	CONSTRAINT "user_email_key" UNIQUE ("email")
);