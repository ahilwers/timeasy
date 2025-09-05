-- Create table for external project connections (GitHub/GitLab/Jira)
create table public.external_connections
(
    id          uuid not null primary key,
    project_id  uuid not null
        constraint fk_external_connections_project
            references public.projects,
    provider    text not null, -- 'github', 'gitlab', 'jira'
    project_ref text not null, -- 'owner/repo' for GitHub/GitLab, 'projectKey' for Jira
    oauth_token text not null, -- encrypted OAuth token
    created_at  timestamp default current_timestamp,
    updated_at  timestamp default current_timestamp,
    constraint unique_project_provider unique (project_id, provider)
);

-- Create table for cached external issues
create table public.external_issues
(
    id            uuid not null primary key,
    project_id    uuid not null
        constraint fk_external_issues_project
            references public.projects,
    provider      text not null, -- 'github', 'gitlab', 'jira'
    key_or_number text not null, -- '#123', 'ABC-123'
    title         text not null,
    state         text not null, -- 'open', 'closed', etc.
    url           text,
    created_at    timestamp default current_timestamp,
    updated_at    timestamp default current_timestamp,
    constraint unique_project_provider_key unique (project_id, provider, key_or_number)
);

-- Add external_issue_id column to time_entries table
alter table public.time_entries 
add column external_issue_id uuid
    constraint fk_time_entries_external_issue
        references public.external_issues;

-- Add pending_external_ref column for offline Flutter app
alter table public.time_entries 
add column pending_external_ref text;

-- Create index for better performance on external issue lookups
create index idx_external_issues_project_provider on public.external_issues(project_id, provider);
create index idx_external_issues_key_or_number on public.external_issues(key_or_number);
create index idx_time_entries_external_issue on public.time_entries(external_issue_id);