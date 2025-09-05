-- Fix missing external integration columns
-- Add columns only if they don't exist

-- Add external_issue_id column to time_entries if it doesn't exist
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'time_entries' 
                   AND column_name = 'external_issue_id') THEN
        -- First check if external_issues table exists
        IF EXISTS (SELECT 1 FROM information_schema.tables 
                   WHERE table_name = 'external_issues' AND table_schema = 'public') THEN
            ALTER TABLE public.time_entries 
            ADD COLUMN external_issue_id uuid
                CONSTRAINT fk_time_entries_external_issue
                    REFERENCES public.external_issues(id);
        ELSE
            -- Add without foreign key constraint if external_issues doesn't exist
            ALTER TABLE public.time_entries 
            ADD COLUMN external_issue_id uuid;
        END IF;
    END IF;
END $$;

-- Add pending_external_ref column if it doesn't exist
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'time_entries' 
                   AND column_name = 'pending_external_ref') THEN
        ALTER TABLE public.time_entries 
        ADD COLUMN pending_external_ref text;
    END IF;
END $$;

-- Create external_issues table if it doesn't exist
CREATE TABLE IF NOT EXISTS public.external_issues
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

-- Create external_connections table if it doesn't exist
CREATE TABLE IF NOT EXISTS public.external_connections
(
    id          uuid not null primary key,
    project_id  uuid not null
        constraint fk_external_connections_project
            references public.projects,
    provider    text not null, -- 'github', 'gitlab', 'jira'
    project_ref text not null, -- 'owner/repo' for GitHub/GitLab, 'projectKey' for Jira
    created_at  timestamp default current_timestamp,
    updated_at  timestamp default current_timestamp,
    constraint unique_project_provider unique (project_id, provider)
);

-- Add user_account_id to external_connections if it doesn't exist
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'external_connections' 
                   AND column_name = 'user_account_id') THEN
        ALTER TABLE public.external_connections 
        ADD COLUMN user_account_id uuid;
    END IF;
END $$;

-- Create user_external_accounts table if it doesn't exist
CREATE TABLE IF NOT EXISTS public.user_external_accounts
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

-- Add foreign key constraint to external_connections.user_account_id if it doesn't exist
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints 
                   WHERE constraint_name = 'fk_external_connections_user_account') THEN
        ALTER TABLE public.external_connections 
        ADD CONSTRAINT fk_external_connections_user_account
            FOREIGN KEY (user_account_id) REFERENCES public.user_external_accounts(id);
    END IF;
END $$;

-- Add foreign key constraint to time_entries.external_issue_id if external_issues exists and constraint doesn't exist
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.tables 
               WHERE table_name = 'external_issues' AND table_schema = 'public') THEN
        IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints 
                       WHERE constraint_name = 'fk_time_entries_external_issue') THEN
            ALTER TABLE public.time_entries 
            ADD CONSTRAINT fk_time_entries_external_issue
                FOREIGN KEY (external_issue_id) REFERENCES public.external_issues(id);
        END IF;
    END IF;
END $$;

-- Create indexes if they don't exist
CREATE INDEX IF NOT EXISTS idx_external_issues_project_provider ON public.external_issues(project_id, provider);
CREATE INDEX IF NOT EXISTS idx_external_issues_key_or_number ON public.external_issues(key_or_number);
CREATE INDEX IF NOT EXISTS idx_time_entries_external_issue_id ON public.time_entries(external_issue_id);
CREATE INDEX IF NOT EXISTS idx_user_external_accounts_user_provider ON public.user_external_accounts(user_id, provider);