package db

const schema = `
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
`

const upsertSQL = `
insert into downloader_jobs (
  id, cache_key, url, status, title, options, progress, resolved, artifact, storage, artifact_expires_at,
  error, error_code, queued_at, started_at, finished_at, download_ms, mux_ms, total_ms, updated_at
) values (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
  $11, $12, $13, $14, $15, $16, $17, $18, $19, now()
)
on conflict (id) do update set
  cache_key = excluded.cache_key,
  url = excluded.url,
  status = excluded.status,
  title = excluded.title,
  options = excluded.options,
  progress = excluded.progress,
  resolved = excluded.resolved,
  artifact = excluded.artifact,
  storage = excluded.storage,
  artifact_expires_at = excluded.artifact_expires_at,
  error = excluded.error,
  error_code = excluded.error_code,
  started_at = excluded.started_at,
  finished_at = excluded.finished_at,
  download_ms = excluded.download_ms,
  mux_ms = excluded.mux_ms,
  total_ms = excluded.total_ms,
  updated_at = now()
`

const loadDoneSQL = `
select id, cache_key, url, status, title, options, progress, resolved, artifact, storage,
  artifact_expires_at, error, error_code, queued_at, started_at, finished_at, download_ms, mux_ms, total_ms
from downloader_jobs
where status = 'done' and artifact <> '' and (artifact_expires_at is null or artifact_expires_at > now())
order by finished_at desc nulls last
limit 10000
`

const loadRunnableSQL = `
select id, cache_key, url, status, title, options, progress, resolved, artifact, storage,
  artifact_expires_at, error, error_code, queued_at, started_at, finished_at, download_ms, mux_ms, total_ms
from downloader_jobs
where status in ('queued', 'running')
order by queued_at asc
limit 10000
`
