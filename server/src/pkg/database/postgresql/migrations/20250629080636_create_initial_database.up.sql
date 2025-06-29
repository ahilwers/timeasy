create table public.teams
(
  id         uuid not null
    primary key,
  name1      text,
  name2      text,
  name3      text
);

create table public.projects
(
  id                   uuid not null
    primary key,
  name                 text,
  user_id              uuid,
  team_id              uuid
    constraint fk_projects_team
      references public.teams,
  color                text           default '#1E90FF'::text,
  deadline             date,
  hourly_rate          numeric(10, 2) default 0.00,
  time_budget          bigint         default 0,
  is_closed            boolean        default false,
  is_active            boolean        default true
);

create table public.time_entries
(
  id          uuid not null
    primary key,
  user_id     uuid,
  project_id  uuid
    constraint fk_time_entries_project
      references public.projects,
  start_time  timestamp,
  end_time    timestamp,
  description text
);

create table public.user_team_assignments
(
  id         bigserial
    primary key,
  user_id    text,
  team_id    uuid
    constraint fk_user_team_assignments_team
      references public.teams,
  roles      varchar(255)
);
