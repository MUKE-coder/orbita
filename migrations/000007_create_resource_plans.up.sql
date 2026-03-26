CREATE TABLE resource_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    max_cpu_cores INTEGER NOT NULL DEFAULT 1,
    max_ram_mb INTEGER NOT NULL DEFAULT 1024,
    max_disk_gb INTEGER NOT NULL DEFAULT 10,
    max_apps INTEGER NOT NULL DEFAULT 5,
    max_databases INTEGER NOT NULL DEFAULT 3,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_resource_plans_deleted_at ON resource_plans(deleted_at);

-- Seed default plans
INSERT INTO resource_plans (id, name, max_cpu_cores, max_ram_mb, max_disk_gb, max_apps, max_databases)
VALUES
    (uuid_generate_v4(), 'Free', 1, 512, 5, 2, 1),
    (uuid_generate_v4(), 'Starter', 2, 2048, 20, 5, 3),
    (uuid_generate_v4(), 'Pro', 4, 8192, 50, 20, 10),
    (uuid_generate_v4(), 'Enterprise', 16, 32768, 200, 100, 50);
