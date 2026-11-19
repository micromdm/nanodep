
CREATE TABLE dep_names (
    name VARCHAR(255) NOT NULL,

    -- OAuth1 Tokens
    consumer_key        TEXT NULL,
	consumer_secret     TEXT NULL,
	access_token        TEXT NULL,
	access_secret       TEXT NULL,
	access_token_expiry TIMESTAMPTZ NULL,

    -- Config
    config_base_url VARCHAR(255) NULL,

    -- Token PKI
    tokenpki_cert_pem         TEXT NULL,
    tokenpki_key_pem          TEXT NULL,
    tokenpki_staging_cert_pem TEXT NULL,
    tokenpki_staging_key_pem  TEXT NULL,

    -- Syncer
    -- From Apple docs: "The string can be up to 1000 characters".
    syncer_cursor VARCHAR(1024) NULL,

    -- Assigner
    assigner_profile_uuid    TEXT NULL,
    assigner_profile_uuid_at TIMESTAMPTZ NULL,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (name),

    CHECK (tokenpki_cert_pem IS NULL OR SUBSTRING(tokenpki_cert_pem FROM 1 FOR 27) = '-----BEGIN CERTIFICATE-----'),
    CHECK (tokenpki_key_pem IS NULL OR SUBSTRING(tokenpki_key_pem FROM 1 FOR  5) = '-----')
);


CREATE  FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_updated_at_on_change
    BEFORE UPDATE
    ON
        dep_names
    FOR EACH ROW
EXECUTE PROCEDURE update_updated_at();

-- Synced DEP devices, keyed by compound (dep_name, serial_number):
-- the same serial may exist under multiple DEP names independently.
-- Nullable columns merge NULL-averse: see devices.sql.
-- All timestamps (including created_at/updated_at) are Unix microseconds.
CREATE TABLE dep_devices (
    serial_number VARCHAR(255) NOT NULL,
    dep_name VARCHAR(255) NOT NULL,
    model VARCHAR(255) NOT NULL,

    asset_tag VARCHAR(255) NULL,
    bluetooth_mac_address VARCHAR(255) NULL,
    color VARCHAR(255) NULL,
    description VARCHAR(255) NULL,
    device_assigned_by VARCHAR(255) NULL,
    device_assigned_date BIGINT NULL,
    device_family VARCHAR(255) NULL,
    eid VARCHAR(255) NULL,
    ethernet_mac_address VARCHAR(255) NULL,
    imei JSONB NULL,
    is_replacement_device BOOLEAN NULL,
    mdm_migration_deadline BIGINT NULL,
    meid JSONB NULL,
    op_date BIGINT NULL,
    op_type VARCHAR(255) NULL,
    os VARCHAR(255) NULL,
    profile_assign_time BIGINT NULL,
    profile_push_time BIGINT NULL,
    profile_status VARCHAR(255) NULL,
    profile_uuid VARCHAR(255) NULL,
    released_by_replacement BOOLEAN NULL,
    response_status VARCHAR(255) NULL,
    wifi_mac_address VARCHAR(255) NULL,

    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,

    PRIMARY KEY (dep_name, serial_number)
);

CREATE INDEX idx_dep_devices_serial ON dep_devices (serial_number);
