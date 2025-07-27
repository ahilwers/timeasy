ALTER TABLE public.teams
ADD COLUMN deleted boolean DEFAULT false;

ALTER TABLE public.projects
ADD COLUMN deleted boolean DEFAULT false;

ALTER TABLE public.time_entries
ADD COLUMN deleted boolean DEFAULT false;

CREATE INDEX idx_teams_deleted ON public.teams (deleted);
CREATE INDEX idx_projects_deleted ON public.projects (deleted);
CREATE INDEX idx_time_entries_deleted ON public.time_entries (deleted);

