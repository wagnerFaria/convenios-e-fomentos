-- name: ListRuleParameters :many
SELECT * FROM rule_parameters
WHERE (sqlc.narg('key')::text IS NULL OR key = sqlc.narg('key'))
ORDER BY key, valid_from DESC;

-- name: GetRuleParameter :one
SELECT * FROM rule_parameters WHERE id = $1;

-- name: GetRuleParameterForUpdate :one
SELECT * FROM rule_parameters WHERE id = $1 FOR UPDATE;

-- name: CreateRuleParameter :one
INSERT INTO rule_parameters (key, value, valid_from, valid_to, note)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateRuleParameter :one
UPDATE rule_parameters
SET value = $2, valid_from = $3, valid_to = $4, note = $5, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteRuleParameter :execrows
DELETE FROM rule_parameters WHERE id = $1;

-- Versões da mesma chave cujo período se sobrepõe ao informado (valid_to nulo = sem fim).
-- name: CountOverlappingRuleParameters :one
SELECT count(*) FROM rule_parameters
WHERE key = $1 AND id <> $2
  AND valid_from <= COALESCE(sqlc.narg('valid_to')::date, 'infinity'::date)
  AND COALESCE(valid_to, 'infinity'::date) >= sqlc.arg('valid_from')::date;

-- Serializa a criação/alteração de versões da mesma chave (a checagem de sobreposição não vê linhas ainda não confirmadas).
-- name: LockRuleParameterKey :exec
SELECT pg_advisory_xact_lock(hashtextextended('rule_parameters:' || $1::text, 0));

-- Valor vigente de cada chave na data.
-- name: ListEffectiveRuleParameters :many
SELECT DISTINCT ON (key) *
FROM rule_parameters
WHERE valid_from <= $1::date AND (valid_to IS NULL OR valid_to >= $1::date)
ORDER BY key, valid_from DESC;

-- Carga dos padrões: não altera o que um administrador já editou.
-- name: InsertDefaultRuleParameter :execrows
INSERT INTO rule_parameters (key, value, valid_from, valid_to, note)
SELECT $1, $2, $3, $4, $5
WHERE NOT EXISTS (SELECT 1 FROM rule_parameters WHERE key = $1);
