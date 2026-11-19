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
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?,
        ?
    ) AS new ON DUPLICATE KEY
UPDATE
    model = new.model,
    asset_tag = COALESCE(new.asset_tag, dep_devices.asset_tag),
    bluetooth_mac_address = COALESCE(
        new.bluetooth_mac_address,
        dep_devices.bluetooth_mac_address
    ),
    color = COALESCE(new.color, dep_devices.color),
    description = COALESCE(new.description, dep_devices.description),
    device_assigned_by = COALESCE(
        new.device_assigned_by,
        dep_devices.device_assigned_by
    ),
    device_assigned_date = COALESCE(
        new.device_assigned_date,
        dep_devices.device_assigned_date
    ),
    device_family = COALESCE(new.device_family, dep_devices.device_family),
    eid = COALESCE(new.eid, dep_devices.eid),
    ethernet_mac_address = COALESCE(
        new.ethernet_mac_address,
        dep_devices.ethernet_mac_address
    ),
    imei = COALESCE(new.imei, dep_devices.imei),
    is_replacement_device = COALESCE(
        new.is_replacement_device,
        dep_devices.is_replacement_device
    ),
    mdm_migration_deadline = COALESCE(
        new.mdm_migration_deadline,
        dep_devices.mdm_migration_deadline
    ),
    meid = COALESCE(new.meid, dep_devices.meid),
    op_date = COALESCE(new.op_date, dep_devices.op_date),
    op_type = COALESCE(new.op_type, dep_devices.op_type),
    os = COALESCE(new.os, dep_devices.os),
    profile_assign_time = COALESCE(
        new.profile_assign_time,
        dep_devices.profile_assign_time
    ),
    profile_push_time = COALESCE(
        new.profile_push_time,
        dep_devices.profile_push_time
    ),
    profile_status = COALESCE(new.profile_status, dep_devices.profile_status),
    profile_uuid = COALESCE(new.profile_uuid, dep_devices.profile_uuid),
    released_by_replacement = COALESCE(
        new.released_by_replacement,
        dep_devices.released_by_replacement
    ),
    response_status = COALESCE(new.response_status, dep_devices.response_status),
    wifi_mac_address = COALESCE(
        new.wifi_mac_address,
        dep_devices.wifi_mac_address
    ),
    created_at = dep_devices.created_at,
    updated_at = new.updated_at;

-- name: ListDevices :many
SELECT
    *
FROM
    dep_devices
ORDER BY
    dep_name,
    serial_number
LIMIT
    ? OFFSET ?;

-- name: ListDevicesByDEP :many
SELECT
    *
FROM
    dep_devices
WHERE
    dep_name IN (sqlc.slice('dep_names'))
ORDER BY
    dep_name,
    serial_number
LIMIT
    ? OFFSET ?;

-- name: ListDevicesBySerial :many
SELECT
    *
FROM
    dep_devices
WHERE
    serial_number IN (sqlc.slice('serials'))
ORDER BY
    dep_name,
    serial_number
LIMIT
    ? OFFSET ?;

-- name: ListDevicesByDEPAndSerial :many
SELECT
    *
FROM
    dep_devices
WHERE
    dep_name IN (sqlc.slice('dep_names'))
    AND serial_number IN (sqlc.slice('serials'))
ORDER BY
    dep_name,
    serial_number
LIMIT
    ? OFFSET ?;

-- name: DeleteDevicesByDEPAndSerials :exec
DELETE FROM
    dep_devices
WHERE
    dep_name = ?
    AND serial_number IN (sqlc.slice('serials'));
