-- Remove the new foreign key constraint
alter table public.external_connections drop column if exists user_account_id;

-- Restore the oauth_token column (data will be lost)
alter table public.external_connections add column oauth_token text;

-- Remove indexes
drop index if exists idx_user_external_accounts_user_provider;

-- Drop the user external accounts table
drop table if exists public.user_external_accounts;