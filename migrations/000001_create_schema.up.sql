-- Table Tenants (Rumah Sakit / External Apps)
CREATE TABLE IF NOT EXISTS tenants (
    id SERIAL PRIMARY KEY,
    key_identifier VARCHAR(100) UNIQUE NOT NULL,
    app_name VARCHAR(100) NOT NULL,
    tenant_name VARCHAR(150) NOT NULL,
    api_key VARCHAR(255) NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table Tickets (Sesi Kendala / Chat)
CREATE TABLE IF NOT EXISTS tickets (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    ticket_code VARCHAR(50) UNIQUE NOT NULL,
    user_id VARCHAR(100) NULL,
    user_name VARCHAR(150) NOT NULL,
    module_name VARCHAR(100) DEFAULT 'Umum',
    topic_title VARCHAR(255) NULL,
    telegram_thread_id BIGINT NULL,
    assigned_programmer VARCHAR(150) NULL,
    status VARCHAR(30) DEFAULT 'open', -- open, escalated, waiting_user, resolved
    csat_rating INT NULL,
    csat_review TEXT NULL,
    diagnostic_info JSONB NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tickets_thread_id ON tickets(telegram_thread_id);
CREATE INDEX IF NOT EXISTS idx_tickets_tenant_status ON tickets(tenant_id, status);

-- Table Messages (Riwayat Pesan & Lampiran)
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    ticket_id INT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    sender_type VARCHAR(20) NOT NULL, -- user, ai, programmer, system
    sender_name VARCHAR(150) NULL,
    message TEXT NOT NULL,
    is_attachment BOOLEAN DEFAULT FALSE,
    attachment_type VARCHAR(20) NULL, -- image, video, document
    attachment_url TEXT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_ticket_id ON messages(ticket_id);
