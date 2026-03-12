-- +goose Up
CREATE TABLE "users" (
	"id" UUID PRIMARY KEY,
	"name" TEXT NOT NULL,
	"email" TEXT NOT NULL,
	"password" TEXT NOT NULL,
	"created_at" TIMESTAMP(3) NOT NULL DEFAULT now(),
	"updated_at" TIMESTAMP(3),
	CONSTRAINT "user_email_key" UNIQUE ("email")
);