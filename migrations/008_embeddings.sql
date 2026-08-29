-- Optional semantic retrieval schema.
-- Requires the pgvector extension to be installed on the PostgreSQL server.
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS memory_embeddings (
    memory_id uuid NOT NULL REFERENCES memories(id),
    model varchar(100) NOT NULL,
    dimensions integer NOT NULL CHECK (dimensions = 768),
    embedding vector(768) NOT NULL,
    source_sha256 char(64) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    PRIMARY KEY (memory_id, model)
);

CREATE INDEX IF NOT EXISTS idx_memory_embeddings_cosine
    ON memory_embeddings USING hnsw (embedding vector_cosine_ops);
