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
