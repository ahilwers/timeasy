DROP INDEX IF EXISTS idx_teams_deleted;
DROP INDEX IF EXISTS idx_projects_deleted;
DROP INDEX IF EXISTS idx_time_entries_deleted;

ALTER TABLE public.teams
DROP COLUMN IF EXISTS deleted;

ALTER TABLE public.projects
DROP COLUMN IF EXISTS deleted;

ALTER TABLE public.time_entries
DROP COLUMN IF EXISTS deleted;
