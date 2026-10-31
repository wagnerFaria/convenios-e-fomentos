-- +goose Up

-- O Analista pode reprovar um item da conferência com uma justificativa. Reprovado não é devolução: o item só vai ao proponente
-- quando o Analista pede a complementação (que o devolve a pendente, como já fazia com os aprovados).
ALTER TABLE project_item_reviews DROP CONSTRAINT project_item_reviews_status_check;
ALTER TABLE project_item_reviews ADD CONSTRAINT project_item_reviews_status_check CHECK (status IN ('pending', 'approved', 'rejected'));
ALTER TABLE project_item_reviews ADD COLUMN note text;
ALTER TABLE project_item_reviews ADD CONSTRAINT project_item_reviews_note_check
    CHECK ((status = 'rejected') = (note IS NOT NULL AND length(btrim(note)) > 0));

-- +goose Down
UPDATE project_item_reviews SET status = 'pending', reviewed_by_id = NULL, reviewed_at = NULL, note = NULL WHERE status = 'rejected';
ALTER TABLE project_item_reviews DROP CONSTRAINT project_item_reviews_note_check;
ALTER TABLE project_item_reviews DROP COLUMN note;
ALTER TABLE project_item_reviews DROP CONSTRAINT project_item_reviews_status_check;
ALTER TABLE project_item_reviews ADD CONSTRAINT project_item_reviews_status_check CHECK (status IN ('pending', 'approved'));
