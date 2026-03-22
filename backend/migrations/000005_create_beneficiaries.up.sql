-- Beneficiaries table
CREATE TABLE IF NOT EXISTS beneficiaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone_number VARCHAR(50),
    national_id VARCHAR(100),
    address TEXT,
    date_of_birth DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Trigger for auto-updating updated_at
CREATE TRIGGER set_beneficiaries_updated_at
    BEFORE UPDATE ON beneficiaries
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Indexes
CREATE INDEX idx_beneficiaries_status ON beneficiaries (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_beneficiaries_full_name ON beneficiaries USING gin (full_name gin_trgm_ops);
CREATE INDEX idx_beneficiaries_email ON beneficiaries (email) WHERE deleted_at IS NULL;
CREATE INDEX idx_beneficiaries_national_id ON beneficiaries (national_id) WHERE deleted_at IS NULL;

-- Program-Beneficiary junction table (many-to-many)
CREATE TABLE IF NOT EXISTS program_beneficiaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    program_id UUID NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
    beneficiary_id UUID NOT NULL REFERENCES beneficiaries(id) ON DELETE CASCADE,
    enrolled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    notes TEXT,
    CONSTRAINT uq_program_beneficiary UNIQUE (program_id, beneficiary_id)
);

CREATE INDEX idx_program_beneficiaries_program ON program_beneficiaries (program_id);
CREATE INDEX idx_program_beneficiaries_beneficiary ON program_beneficiaries (beneficiary_id);
