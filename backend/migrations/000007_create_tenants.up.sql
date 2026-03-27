-- Tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone_number VARCHAR(50),
    national_id VARCHAR(100),
    address_street VARCHAR(255),
    address_city VARCHAR(255),
    address_state_or_region VARCHAR(255),
    address_postal_code VARCHAR(20),
    address_country VARCHAR(100) DEFAULT 'Philippines',
    landlord_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Trigger for auto-updating updated_at
CREATE TRIGGER set_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Indexes
CREATE INDEX idx_tenants_landlord_id ON tenants (landlord_id);
CREATE UNIQUE INDEX idx_tenants_email ON tenants (email) WHERE email IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_tenants_is_active ON tenants (is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_deleted_at ON tenants (deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_tenants_full_name ON tenants USING gin (full_name gin_trgm_ops);
