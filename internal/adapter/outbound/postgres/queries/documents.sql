-- name: CreateDocumentSet :one
INSERT INTO document_sets (user_id, original_filename, status, object_key)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetDocumentSetByIDAndUser :one
SELECT * FROM document_sets
WHERE id = $1 AND user_id = $2;

-- name: UpdateDocumentSetStatus :one
UPDATE document_sets
SET status = $1, error = $2, updated_at = now()
WHERE id = $3
RETURNING *;

-- name: InsertDocument :one
INSERT INTO documents (set_id, user_id, filename, content_type, size_bytes, object_key)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListDocumentsByUser :many
SELECT * FROM documents
WHERE user_id = $1
ORDER BY created_at ASC;

-- name: GetDocumentByIDAndUser :one
SELECT * FROM documents
WHERE id = $1 AND user_id = $2;

-- name: DeleteDocumentsBySetID :exec
DELETE FROM documents WHERE set_id = $1;

-- name: CountDocumentsBySetID :one
SELECT COUNT(*) FROM documents WHERE set_id = $1;
