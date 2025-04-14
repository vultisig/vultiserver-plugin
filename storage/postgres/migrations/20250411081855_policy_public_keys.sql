-- +goose Up
-- +goose StatementBegin

-- drop contraints so table can be safely deleted
ALTER TABLE time_triggers DROP CONSTRAINT time_triggers_policy_id_fkey;
ALTER TABLE transaction_history DROP CONSTRAINT transaction_history_policy_id_fkey;
ALTER TABLE transaction_history DROP CONSTRAINT fk_policy;
DROP TABLE IF EXISTS plugin_policies;

-- recreate table with proper field order and new fields (public keys, chain_id)
CREATE TABLE plugin_policies (
    id UUID PRIMARY KEY,
    public_key_ecdsa TEXT NOT NULL,
    public_key_eddsa TEXT NOT NULL,
    chain_id VARCHAR(255) NOT NULL,
    plugin_version TEXT NOT NULL,
    policy_version TEXT NOT NULL,
    plugin_type plugin_type NOT NULL,
    is_ecdsa BOOLEAN DEFAULT TRUE,
    chain_code_hex TEXT NOT NULL,
    derive_path TEXT NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    signature TEXT NOT NULL,
    policy JSONB NOT NULL
);

-- faster lookups on type as it is the primary relation we use
CREATE INDEX idx_plugin_policies_plugin_type ON plugin_policies(plugin_type);

-- reapply constraints (fk_policy is redundant, skip it)
ALTER TABLE time_triggers
ADD CONSTRAINT time_triggers_policy_id_fkey
FOREIGN KEY (policy_id)
REFERENCES plugin_policies(id);

ALTER TABLE transaction_history
ADD CONSTRAINT transaction_history_policy_id_fkey
FOREIGN KEY (policy_id)
REFERENCES plugin_policies(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- drop contraints so table can be safely deleted
ALTER TABLE time_triggers DROP CONSTRAINT time_triggers_policy_id_fkey;
ALTER TABLE transaction_history DROP CONSTRAINT transaction_history_policy_id_fkey;
DROP TABLE IF EXISTS plugin_policies;

-- create table as it was before this migration
CREATE TABLE plugin_policies (
    id UUID PRIMARY KEY,
    public_key TEXT NOT NULL,
    plugin_version TEXT NOT NULL,
    policy_version TEXT NOT NULL,
    plugin_type plugin_type NOT NULL,
    signature TEXT NOT NULL,
    policy JSONB NOT NULL,
    is_ecdsa BOOLEAN DEFAULT TRUE,
    chain_code_hex TEXT NOT NULL,
    derive_path TEXT NOT NULL,
    active BOOLEAN DEFAULT TRUE
);

-- add constraints as they were before this migration
ALTER TABLE time_triggers
ADD CONSTRAINT time_triggers_policy_id_fkey
FOREIGN KEY (policy_id)
REFERENCES plugin_policies(id);

ALTER TABLE transaction_history
ADD CONSTRAINT transaction_history_policy_id_fkey
FOREIGN KEY (policy_id)
REFERENCES plugin_policies(id);

ALTER TABLE transaction_history
ADD CONSTRAINT fk_policy
FOREIGN KEY (policy_id)
REFERENCES plugin_policies(id);

-- +goose StatementEnd
