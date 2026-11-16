CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    login VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'CLIENT',
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Demo user for local testing (login: demo, password: demo1234). Change in production.
INSERT INTO users (id, login, password_hash, role, status)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'demo',
    '$2a$10$JqsY8GbOzZrwSAGpaToKNeLOyquZaQ.5VCfaXc6wLKIOve4pDSl16',
    'CLIENT',
    'ACTIVE'
)
ON CONFLICT (login) DO NOTHING;
