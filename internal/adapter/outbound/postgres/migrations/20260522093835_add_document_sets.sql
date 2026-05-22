-- Create "document_sets" table
CREATE TABLE "public"."document_sets" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "user_id" uuid NOT NULL, "original_filename" character varying(255) NOT NULL, "status" character varying(20) NOT NULL, "object_key" text NOT NULL, "error" text NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"), CONSTRAINT "fk_document_sets_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_document_sets_user_id" to table: "document_sets"
CREATE INDEX "idx_document_sets_user_id" ON "public"."document_sets" ("user_id");
