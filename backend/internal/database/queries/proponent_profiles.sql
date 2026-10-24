-- name: GetProponentProfile :one
SELECT * FROM proponent_profiles WHERE admin_id = $1;

-- name: UpsertProponentProfile :one
INSERT INTO proponent_profiles (admin_id, phones, emails, uf, municipality_ibge, job_title, personal)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (admin_id) DO UPDATE
SET phones = EXCLUDED.phones, emails = EXCLUDED.emails, uf = EXCLUDED.uf,
    municipality_ibge = EXCLUDED.municipality_ibge, job_title = EXCLUDED.job_title, personal = EXCLUDED.personal, updated_at = now()
RETURNING *;
