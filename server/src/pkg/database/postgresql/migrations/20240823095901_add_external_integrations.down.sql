-- Remove indexes
drop index if exists idx_time_entries_external_issue;
drop index if exists idx_external_issues_key_or_number;
drop index if exists idx_external_issues_project_provider;

-- Remove columns from time_entries table
alter table public.time_entries drop column if exists pending_external_ref;
alter table public.time_entries drop column if exists external_issue_id;

-- Drop external tables
drop table if exists public.external_issues;
drop table if exists public.external_connections;