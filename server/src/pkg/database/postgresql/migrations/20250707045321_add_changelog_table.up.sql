CREATE TABLE public.change_log (
  id BIGSERIAL PRIMARY KEY,
  entity_type TEXT NOT NULL,
  entity_id  UUID NOT NULL,
  operation VARCHAR(10) NOT NULL CHECK (operation IN ('Created','Updated','Deleted')),
  changed_by_user UUID NOT NULL,
  changed_by_client TEXT NULL,
  changed_at TIMESTAMP NOT NULL DEFAULT now()
);
CREATE INDEX idx_changelog_entity ON change_log(entity_type, entity_id);
CREATE INDEX idx_changelog_changeid ON change_log(id);
CREATE INDEX idx_changelog_id_type_client ON change_log(id, entity_type, changed_by_client);
