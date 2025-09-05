-- Rollback external integration columns and tables

-- Drop indexes
DROP INDEX IF EXISTS idx_user_external_accounts_user_provider;
DROP INDEX IF EXISTS idx_time_entries_external_issue_id;
DROP INDEX IF EXISTS idx_external_issues_key_or_number;
DROP INDEX IF EXISTS idx_external_issues_project_provider;

-- Remove foreign key constraints
ALTER TABLE public.time_entries DROP CONSTRAINT IF EXISTS fk_time_entries_external_issue;
ALTER TABLE public.external_connections DROP CONSTRAINT IF EXISTS fk_external_connections_user_account;

-- Drop columns
ALTER TABLE public.time_entries DROP COLUMN IF EXISTS pending_external_ref;
ALTER TABLE public.time_entries DROP COLUMN IF EXISTS external_issue_id;
ALTER TABLE public.external_connections DROP COLUMN IF EXISTS user_account_id;

-- Drop tables
DROP TABLE IF EXISTS public.user_external_accounts;
DROP TABLE IF EXISTS public.external_connections;
DROP TABLE IF EXISTS public.external_issues;