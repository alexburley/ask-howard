-- Modify "users" table
ALTER TABLE "public"."users" ADD COLUMN "last_login_at" timestamptz NULL;
