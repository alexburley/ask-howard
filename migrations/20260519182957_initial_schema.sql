-- Create "documents" table
CREATE TABLE "public"."documents" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "title" character varying(255) NOT NULL, "document_type" character varying(100) NULL, "file_url" text NULL, "uploaded_at" timestamptz NOT NULL DEFAULT now(), "approx_date" date NULL, "source_notes" text NULL, "ocr_status" character varying(50) NOT NULL DEFAULT 'pending', "analysis_status" character varying(50) NOT NULL DEFAULT 'pending', "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"));
-- Create "claims" table
CREATE TABLE "public"."claims" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "subject_type" character varying(100) NOT NULL, "subject_id" uuid NOT NULL, "claim_type" character varying(100) NOT NULL, "value" text NOT NULL, "source_document_id" uuid NULL, "evidence_text" text NULL, "confidence" double precision NOT NULL DEFAULT 0, "status" character varying(50) NOT NULL DEFAULT 'suggested', "created_by" character varying(255) NOT NULL DEFAULT 'system', "reviewed_at" timestamptz NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"), CONSTRAINT "fk_claims_source_document" FOREIGN KEY ("source_document_id") REFERENCES "public"."documents" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "idx_claims_status" to table: "claims"
CREATE INDEX "idx_claims_status" ON "public"."claims" ("status");
-- Create index "idx_claims_subject" to table: "claims"
CREATE INDEX "idx_claims_subject" ON "public"."claims" ("subject_type", "subject_id");
-- Create "extracted_texts" table
CREATE TABLE "public"."extracted_texts" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "document_id" uuid NOT NULL, "raw_text" text NULL, "corrected_text" text NULL, "ocr_confidence" double precision NULL, "extraction_method" character varying(50) NOT NULL, "created_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"), CONSTRAINT "fk_extracted_texts_document" FOREIGN KEY ("document_id") REFERENCES "public"."documents" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_extracted_texts_document_id" to table: "extracted_texts"
CREATE INDEX "idx_extracted_texts_document_id" ON "public"."extracted_texts" ("document_id");
-- Create "persons" table
CREATE TABLE "public"."persons" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "display_name" character varying(255) NOT NULL, "birth_date" date NULL, "death_date" date NULL, "notes" text NULL, "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"));
-- Create "person_mentions" table
CREATE TABLE "public"."person_mentions" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "document_id" uuid NOT NULL, "text_as_written" character varying(255) NOT NULL, "normalised_name" character varying(255) NULL, "page_number" integer NULL, "bounding_box" jsonb NULL, "confidence" double precision NOT NULL DEFAULT 0, "linked_person_id" uuid NULL, "created_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"), CONSTRAINT "fk_person_mentions_document" FOREIGN KEY ("document_id") REFERENCES "public"."documents" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "fk_person_mentions_person" FOREIGN KEY ("linked_person_id") REFERENCES "public"."persons" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "idx_person_mentions_document_id" to table: "person_mentions"
CREATE INDEX "idx_person_mentions_document_id" ON "public"."person_mentions" ("document_id");
-- Create index "idx_person_mentions_linked_person_id" to table: "person_mentions"
CREATE INDEX "idx_person_mentions_linked_person_id" ON "public"."person_mentions" ("linked_person_id");
-- Create "relationships" table
CREATE TABLE "public"."relationships" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "from_person_id" uuid NOT NULL, "to_person_id" uuid NOT NULL, "relationship_type" character varying(100) NOT NULL, "confidence" double precision NOT NULL DEFAULT 0, "status" character varying(50) NOT NULL DEFAULT 'suggested', "created_at" timestamptz NOT NULL DEFAULT now(), "updated_at" timestamptz NOT NULL DEFAULT now(), PRIMARY KEY ("id"), CONSTRAINT "fk_relationships_from_person" FOREIGN KEY ("from_person_id") REFERENCES "public"."persons" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "fk_relationships_to_person" FOREIGN KEY ("to_person_id") REFERENCES "public"."persons" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_relationships_from_person_id" to table: "relationships"
CREATE INDEX "idx_relationships_from_person_id" ON "public"."relationships" ("from_person_id");
-- Create index "idx_relationships_to_person_id" to table: "relationships"
CREATE INDEX "idx_relationships_to_person_id" ON "public"."relationships" ("to_person_id");
