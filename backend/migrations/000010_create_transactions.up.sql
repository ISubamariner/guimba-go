CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    debt_id UUID NOT NULL REFERENCES debts(id) ON DELETE RESTRICT,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    landlord_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    recorded_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    transaction_type VARCHAR(20) NOT NULL,
    amount DECIMAL(14,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'PHP',
    payment_method VARCHAR(20) NOT NULL,
    transaction_date DATE NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    receipt_number VARCHAR(100),
    reference_number VARCHAR(100),
    is_verified BOOLEAN NOT NULL DEFAULT false,
    verified_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique partial index: reference_number per debt (when not null)
CREATE UNIQUE INDEX idx_transactions_debt_reference ON transactions(debt_id, reference_number) WHERE reference_number IS NOT NULL;

-- Indexes
CREATE INDEX idx_transactions_debt_id ON transactions(debt_id);
CREATE INDEX idx_transactions_tenant_id ON transactions(tenant_id);
CREATE INDEX idx_transactions_landlord_id ON transactions(landlord_id);
CREATE INDEX idx_transactions_type ON transactions(transaction_type);

-- Auto-update updated_at trigger
CREATE TRIGGER set_transactions_updated_at
    BEFORE UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
