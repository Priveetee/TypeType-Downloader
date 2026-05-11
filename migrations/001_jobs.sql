create table if not exists downloader_jobs (
  id text primary key,
  cache_key text not null,
  url text not null,
  status text not null,
  title text not null default '',
  options jsonb not null,
  progress jsonb not null,
  resolved jsonb,
  artifact text not null default '',
  storage text not null default '',
  artifact_expires_at timestamptz,
  error text,
  error_code text,
  queued_at timestamptz not null,
  started_at timestamptz,
  finished_at timestamptz,
  download_ms bigint,
  mux_ms bigint,
  total_ms bigint,
  updated_at timestamptz not null default now()
);

create index if not exists downloader_jobs_cache_key_idx on downloader_jobs(cache_key);
create index if not exists downloader_jobs_status_idx on downloader_jobs(status);
alter table downloader_jobs add column if not exists artifact_expires_at timestamptz;
