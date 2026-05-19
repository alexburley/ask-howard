schema "public" {}

table "users" {
  schema = schema.public

  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "email" {
    null = false
    type = varchar(255)
  }
  column "password_hash" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }

  index "idx_users_email" {
    unique  = true
    columns = [column.email]
  }
}

table "documents" {
  schema = schema.public

  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "title" {
    null = false
    type = varchar(255)
  }
  column "document_type" {
    null = true
    type = varchar(100)
  }
  column "file_url" {
    null = true
    type = text
  }
  column "uploaded_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "approx_date" {
    null = true
    type = date
  }
  column "source_notes" {
    null = true
    type = text
  }
  column "ocr_status" {
    null    = false
    type    = varchar(50)
    default = "pending"
  }
  column "analysis_status" {
    null    = false
    type    = varchar(50)
    default = "pending"
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }
}

table "extracted_texts" {
  schema = schema.public

  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "document_id" {
    null = false
    type = uuid
  }
  column "raw_text" {
    null = true
    type = text
  }
  column "corrected_text" {
    null = true
    type = text
  }
  column "ocr_confidence" {
    null = true
    type = double_precision
  }
  column "extraction_method" {
    null = false
    type = varchar(50)
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }

  foreign_key "fk_extracted_texts_document" {
    columns     = [column.document_id]
    ref_columns = [table.documents.column.id]
    on_delete   = CASCADE
  }

  index "idx_extracted_texts_document_id" {
    columns = [column.document_id]
  }
}

table "persons" {
  schema = schema.public

  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "display_name" {
    null = false
    type = varchar(255)
  }
  column "birth_date" {
    null = true
    type = date
  }
  column "death_date" {
    null = true
    type = date
  }
  column "notes" {
    null = true
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }
}

table "person_mentions" {
  schema = schema.public

  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "document_id" {
    null = false
    type = uuid
  }
  column "text_as_written" {
    null = false
    type = varchar(255)
  }
  column "normalised_name" {
    null = true
    type = varchar(255)
  }
  column "page_number" {
    null = true
    type = integer
  }
  column "bounding_box" {
    null = true
    type = jsonb
  }
  column "confidence" {
    null    = false
    type    = double_precision
    default = 0
  }
  column "linked_person_id" {
    null = true
    type = uuid
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }

  foreign_key "fk_person_mentions_document" {
    columns     = [column.document_id]
    ref_columns = [table.documents.column.id]
    on_delete   = CASCADE
  }

  foreign_key "fk_person_mentions_person" {
    columns     = [column.linked_person_id]
    ref_columns = [table.persons.column.id]
    on_delete   = SET_NULL
  }

  index "idx_person_mentions_document_id" {
    columns = [column.document_id]
  }

  index "idx_person_mentions_linked_person_id" {
    columns = [column.linked_person_id]
  }
}

table "claims" {
  schema = schema.public

  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "subject_type" {
    null = false
    type = varchar(100)
  }
  column "subject_id" {
    null = false
    type = uuid
  }
  column "claim_type" {
    null = false
    type = varchar(100)
  }
  column "value" {
    null = false
    type = text
  }
  column "source_document_id" {
    null = true
    type = uuid
  }
  column "evidence_text" {
    null = true
    type = text
  }
  column "confidence" {
    null    = false
    type    = double_precision
    default = 0
  }
  column "status" {
    null    = false
    type    = varchar(50)
    default = "suggested"
  }
  column "created_by" {
    null    = false
    type    = varchar(255)
    default = "system"
  }
  column "reviewed_at" {
    null = true
    type = timestamptz
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }

  foreign_key "fk_claims_source_document" {
    columns     = [column.source_document_id]
    ref_columns = [table.documents.column.id]
    on_delete   = SET_NULL
  }

  index "idx_claims_subject" {
    columns = [column.subject_type, column.subject_id]
  }

  index "idx_claims_status" {
    columns = [column.status]
  }
}

table "relationships" {
  schema = schema.public

  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "from_person_id" {
    null = false
    type = uuid
  }
  column "to_person_id" {
    null = false
    type = uuid
  }
  column "relationship_type" {
    null = false
    type = varchar(100)
  }
  column "confidence" {
    null    = false
    type    = double_precision
    default = 0
  }
  column "status" {
    null    = false
    type    = varchar(50)
    default = "suggested"
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }

  foreign_key "fk_relationships_from_person" {
    columns     = [column.from_person_id]
    ref_columns = [table.persons.column.id]
    on_delete   = CASCADE
  }

  foreign_key "fk_relationships_to_person" {
    columns     = [column.to_person_id]
    ref_columns = [table.persons.column.id]
    on_delete   = CASCADE
  }

  index "idx_relationships_from_person_id" {
    columns = [column.from_person_id]
  }

  index "idx_relationships_to_person_id" {
    columns = [column.to_person_id]
  }
}
