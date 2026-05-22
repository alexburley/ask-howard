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
