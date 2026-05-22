-- Create "documents" table
CREATE TABLE "public"."documents" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "set_id" uuid NOT NULL, "user_id" uuid NOT NULL, "filename" character varying(255) NOT NULL, "content_type" character varying(100) NOT NULL, "size_bytes" bigint NOT NULL, "object_key" text NOT NULL, "canvas_x" double precision NULL, "canvas_y" double precision NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"), CONSTRAINT "fk_documents_set" FOREIGN KEY ("set_id") REFERENCES "public"."document_sets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "fk_documents_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_documents_set_id" to table: "documents"
CREATE INDEX "idx_documents_set_id" ON "public"."documents" ("set_id");
-- Create index "idx_documents_user_id" to table: "documents"
CREATE INDEX "idx_documents_user_id" ON "public"."documents" ("user_id");
