-- Upsert a single synced device, NULL-averse: NULL incoming fields
-- preserve the stored values (COALESCE), identity columns overwrite.
-- Go loops over devices in slice order; wrap the batch in a transaction.
-- name: UpsertDevice :exec
INSERT INTO
    dep_devices (
        serial_number,
        dep_name,
        model,
        asset_tag,
        bluetooth_mac_address,
        color,
        description,
        device_assigned_by,
        device_assigned_date,
        device_family,
        eid,
        ethernet_mac_address,
        imei,
        is_replacement_device,
        mdm_migration_deadline,
        meid,
        op_date,
        op_type,
        os,
        profile_assign_time,
        profile_push_time,
        profile_status,
        profile_uuid,
        released_by_replacement,
        response_status,
        wifi_mac_address,
        created_at,
        updated_at
    )
VALUES
    (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        $16,
        $17,
        $18,
        $19,
        $20,
        $21,
        $22,
        $23,
        $24,
        $25,
        $26,
        $27,
        $28
    ) ON CONFLICT (dep_name, serial_number) DO
UPDATE
SET
    model = excluded.model,
    asset_tag = COALESCE(excluded.asset_tag, dep_devices.asset_tag),
    bluetooth_mac_address = COALESCE(
        excluded.bluetooth_mac_address,
        dep_devices.bluetooth_mac_address
    ),
    color = COALESCE(excluded.color, dep_devices.color),
    description = COALESCE(excluded.description, dep_devices.description),
    device_assigned_by = COALESCE(
        excluded.device_assigned_by,
        dep_devices.device_assigned_by
    ),
    device_assigned_date = COALESCE(
        excluded.device_assigned_date,
        dep_devices.device_assigned_date
    ),
    device_family = COALESCE(
        excluded.device_family,
        dep_devices.device_family
    ),
    eid = COALESCE(excluded.eid, dep_devices.eid),
    ethernet_mac_address = COALESCE(
        excluded.ethernet_mac_address,
        dep_devices.ethernet_mac_address
    ),
    imei = COALESCE(excluded.imei, dep_devices.imei),
    is_replacement_device = COALESCE(
        excluded.is_replacement_device,
        dep_devices.is_replacement_device
    ),
    mdm_migration_deadline = COALESCE(
        excluded.mdm_migration_deadline,
        dep_devices.mdm_migration_deadline
    ),
    meid = COALESCE(excluded.meid, dep_devices.meid),
    op_date = COALESCE(excluded.op_date, dep_devices.op_date),
    op_type = COALESCE(excluded.op_type, dep_devices.op_type),
    os = COALESCE(excluded.os, dep_devices.os),
    profile_assign_time = COALESCE(
        excluded.profile_assign_time,
        dep_devices.profile_assign_time
    ),
    profile_push_time = COALESCE(
        excluded.profile_push_time,
        dep_devices.profile_push_time
    ),
    profile_status = COALESCE(
        excluded.profile_status,
        dep_devices.profile_status
    ),
    profile_uuid = COALESCE(excluded.profile_uuid, dep_devices.profile_uuid),
    released_by_replacement = COALESCE(
        excluded.released_by_replacement,
        dep_devices.released_by_replacement
    ),
    response_status = COALESCE(
        excluded.response_status,
        dep_devices.response_status
    ),
    wifi_mac_address = COALESCE(
        excluded.wifi_mac_address,
        dep_devices.wifi_mac_address
    ),
    created_at = dep_devices.created_at,
    updated_at = excluded.updated_at;

-- name: ListDevices :many
SELECT
    *
FROM
    dep_devices
ORDER BY
    dep_name,
    serial_number
LIMIT
    $1 OFFSET $2;

-- name: ListDevicesByDEP :many
SELECT
    *
FROM
    dep_devices
WHERE
    dep_name = ANY(sqlc.arg('dep_names')::varchar [])
ORDER BY
    dep_name,
    serial_number
LIMIT
    $1 OFFSET $2;

-- name: ListDevicesBySerial :many
SELECT
    *
FROM
    dep_devices
WHERE
    serial_number = ANY(sqlc.arg('serials')::varchar [])
ORDER BY
    dep_name,
    serial_number
LIMIT
    $1 OFFSET $2;

-- name: ListDevicesByDEPAndSerial :many
SELECT
    *
FROM
    dep_devices
WHERE
    dep_name = ANY(sqlc.arg('dep_names')::varchar [])
    AND serial_number = ANY(sqlc.arg('serials')::varchar [])
ORDER BY
    dep_name,
    serial_number
LIMIT
    $1 OFFSET $2;

-- name: DeleteDevicesByDEPAndSerials :exec
DELETE FROM
    dep_devices
WHERE
    dep_name = $1
    AND serial_number = ANY(sqlc.arg('serials')::varchar []);
