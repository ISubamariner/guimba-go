-- Properties table
CREATE TABLE IF NOT EXISTS properties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    property_code VARCHAR(100) NOT NULL,
    address_street VARCHAR(255),
    address_city VARCHAR(255),
    address_state_or_region VARCHAR(255),
    address_postal_code VARCHAR(20),
    address_country VARCHAR(100) DEFAULT 'Philippines',
    geojson_coordinates TEXT,
    property_type VARCHAR(50) NOT NULL DEFAULT 'LAND',
    size_in_acres DECIMAL(12,4),
    size_in_sqm DECIMAL(12,4) NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    is_available_for_rent BOOLEAN NOT NULL DEFAULT true,
    is_active BOOLEAN NOT NULL DEFAULT true,
    monthly_rent_amount DECIMAL(12,2),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Trigger for auto-updating updated_at
CREATE TRIGGER set_properties_updated_at
    BEFORE UPDATE ON properties
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Indexes
CREATE UNIQUE INDEX idx_properties_property_code ON properties (property_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_properties_owner_id ON properties (owner_id);
CREATE INDEX idx_properties_is_active ON properties (is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_properties_property_type ON properties (property_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_properties_deleted_at ON properties (deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_properties_name ON properties USING gin (name gin_trgm_ops);
