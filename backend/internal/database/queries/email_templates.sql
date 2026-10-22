-- name: GetEmailTemplate :one
SELECT * FROM email_templates WHERE key = $1;

-- name: ListEmailTemplates :many
SELECT * FROM email_templates ORDER BY key;

-- name: UpsertEmailTemplate :one
INSERT INTO email_templates (key, sender_name, sender_email, reply_to, subject, body_html)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (key) DO UPDATE SET sender_name = EXCLUDED.sender_name, sender_email = EXCLUDED.sender_email,
  reply_to = EXCLUDED.reply_to, subject = EXCLUDED.subject, body_html = EXCLUDED.body_html, updated_at = now()
RETURNING *;

-- name: InsertEmailTemplateIfMissing :exec
INSERT INTO email_templates (key, sender_name, sender_email, reply_to, subject, body_html)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (key) DO NOTHING;
