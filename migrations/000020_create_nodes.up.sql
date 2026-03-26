CREATE TABLE nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    ip VARCHAR(45) NOT NULL,
    ssh_port INTEGER NOT NULL DEFAULT 22,
    ssh_key_id UUID,
    role VARCHAR(50) NOT NULL DEFAULT 'worker',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    labels JSONB NOT NULL DEFAULT '{}',
    cpu_cores INTEGER NOT NULL DEFAULT 0,
    ram_mb INTEGER NOT NULL DEFAULT 0,
    disk_gb INTEGER NOT NULL DEFAULT 0,
    docker_version VARCHAR(50),
    swarm_node_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_nodes_status ON nodes(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_nodes_deleted_at ON nodes(deleted_at);
