CREATE TABLE cron_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    schedule VARCHAR(100) NOT NULL,
    image VARCHAR(512) NOT NULL,
    command TEXT,
    env_config TEXT,
    timeout INTEGER NOT NULL DEFAULT 3600,
    concurrency_policy VARCHAR(20) NOT NULL DEFAULT 'forbid',
    max_retries INTEGER NOT NULL DEFAULT 0,
    cpu_limit INTEGER NOT NULL DEFAULT 0,
    memory_limit INTEGER NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_cron_jobs_org_id ON cron_jobs(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_cron_jobs_deleted_at ON cron_jobs(deleted_at);

CREATE TABLE cron_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cron_job_id UUID NOT NULL REFERENCES cron_jobs(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL DEFAULT 'running',
    exit_code INTEGER,
    log_snippet TEXT,
    duration_ms INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cron_runs_cron_job_id ON cron_runs(cron_job_id);
CREATE INDEX idx_cron_runs_status ON cron_runs(status);
