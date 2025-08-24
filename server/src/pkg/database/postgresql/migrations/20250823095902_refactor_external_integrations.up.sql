-- Create table for user external accounts (replaces external_connections)
create table public.user_external_accounts
(
    id          uuid not null primary key,
    user_id     uuid not null,
    provider    text not null, -- 'github', 'gitlab', 'jira'
    account_name text not null, -- Display name for the account
    oauth_token text not null, -- encrypted OAuth token
    base_url    text, -- For GitLab self-hosted or Jira instances
    created_at  timestamp default current_timestamp,
    updated_at  timestamp default current_timestamp,
    constraint unique_user_provider_account unique (user_id, provider, account_name)
);

-- Update external_connections to reference user accounts instead of storing tokens
alter table public.external_connections 
drop column oauth_token;

alter table public.external_connections 
add column user_account_id uuid
    constraint fk_external_connections_user_account
        references public.user_external_accounts(id);

-- Create index for better performance
create index idx_user_external_accounts_user_provider on public.user_external_accounts(user_id, provider);

-- Migrate existing data (this will require manual intervention in production)
-- For now, we'll just add the new structure and existing connections will need to be reconfigured