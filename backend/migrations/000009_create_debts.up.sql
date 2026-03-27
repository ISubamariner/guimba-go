CREATE TABLE IF NOT EXISTS debts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    landlord_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    property_id UUID REFERENCES properties(id) ON DELETE SET NULL,
    debt_type VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    original_amount DECIMAL(14,2) NOT NULL,
    original_currency VARCHAR(3) NOT NULL DEFAULT 'PHP',
    amount_paid DECIMAL(14,2) NOT NULL DEFAULT 0,
    amount_paid_currency VARCHAR(3) NOT NULL DEFAULT 'PHP',
    due_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Indexes
CREATE INDEX idx_debts_tenant_id ON debts(tenant_id);
CREATE INDEX idx_debts_landlord_id ON debts(landlord_id);
CREATE INDEX idx_debts_property_id ON debts(property_id);
CREATE INDEX idx_debts_status ON debts(status);
CREATE INDEX idx_debts_due_date_active ON debts(due_date) WHERE status NOT IN ('PAID', 'CANCELLED') AND deleted_at IS NULL;
CREATE INDEX idx_debts_deleted_at ON debts(deleted_at);

-- Auto-update updated_at trigger
CREATE TRIGGER set_debts_updated_at
    BEFORE UPDATE ON debts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
