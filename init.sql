CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS clients (
    client_id varchar(255) PRIMARY KEY,
    hostname text NOT NULL DEFAULT '',
    os_info text NOT NULL DEFAULT '',
    description text NOT NULL DEFAULT '',
    first_seen_at timestamptz NOT NULL DEFAULT NOW(),
    last_seen_at timestamptz NOT NULL DEFAULT NOW(),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS memories (
    id varchar(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    client_id varchar(255) NOT NULL REFERENCES clients(client_id),
    category varchar(100) NOT NULL DEFAULT 'knowledge',
    title text NOT NULL,
    content_l0 text NOT NULL,
    content_l1 text,
    content_l2 text,
    tags text[] NOT NULL DEFAULT '{}',
    current_level integer NOT NULL DEFAULT 0 CHECK (current_level BETWEEN 0 AND 3),
    valid_from timestamptz NOT NULL DEFAULT NOW(),
    valid_to timestamptz,
    tx_created_at timestamptz NOT NULL DEFAULT NOW(),
    tx_invalidated_at timestamptz,
    status varchar(20) NOT NULL DEFAULT 'ACTIVE',
    superseded_by varchar(255),
    version integer NOT NULL DEFAULT 1,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS knowledge_nodes (
    id varchar(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    name text NOT NULL UNIQUE,
    node_type varchar(100) NOT NULL DEFAULT 'concept',
    summary text,
    attributes jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS knowledge_edges (
    id varchar(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    source_node_id varchar(255) NOT NULL REFERENCES knowledge_nodes(id),
    target_node_id varchar(255) NOT NULL REFERENCES knowledge_nodes(id),
    relation_type varchar(100) NOT NULL,
    weight double precision NOT NULL DEFAULT 1.0,
    evidence_memory_id varchar(255) REFERENCES memories(id),
    valid_from timestamptz NOT NULL DEFAULT NOW(),
    valid_to timestamptz,
    created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE VIEW v_active_memories AS
SELECT * FROM memories
WHERE status = 'ACTIVE' AND valid_to IS NULL;

CREATE OR REPLACE VIEW v_active_knowledge_graph AS
SELECT
    e.id AS edge_id,
    s.name AS source_name,
    s.node_type AS source_type,
    e.relation_type,
    t.name AS target_name,
    t.node_type AS target_type,
    e.weight,
    e.valid_from,
    e.valid_to,
    e.evidence_memory_id
FROM knowledge_edges e
JOIN knowledge_nodes s ON s.id = e.source_node_id
JOIN knowledge_nodes t ON t.id = e.target_node_id
WHERE e.valid_to IS NULL;

CREATE INDEX IF NOT EXISTS idx_memories_active_category ON memories(category, updated_at DESC) WHERE status = 'ACTIVE';
CREATE INDEX IF NOT EXISTS idx_memories_title_trgm ON memories USING gin(title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_memories_content_l0_trgm ON memories USING gin(content_l0 gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_memories_metadata ON memories USING gin(metadata);
CREATE INDEX IF NOT EXISTS idx_memories_tags ON memories USING gin(tags);
CREATE INDEX IF NOT EXISTS idx_edges_active ON knowledge_edges(valid_to);
