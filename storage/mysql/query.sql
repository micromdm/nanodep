-- name: GetConfigBaseURL :one
SELECT config_base_url FROM dep_names WHERE name = ?;

-- name: GetSyncerCursor :one
SELECT syncer_cursor FROM dep_names WHERE name = ?;

-- name: GetCurrentKeypair :one
SELECT
  tokenpki_cert_pem,
  tokenpki_key_pem
FROM
  dep_names
WHERE
  name = ?;

-- name: GetStagingKeypair :one
SELECT
  tokenpki_staging_cert_pem,
  tokenpki_staging_key_pem
FROM
  dep_names
WHERE
  name = ?;

-- name: UpstageKeypair :exec
UPDATE
  dep_names
SET
  tokenpki_cert_pem = tokenpki_staging_cert_pem,
  tokenpki_key_pem = tokenpki_staging_key_pem
WHERE
  name = ?;

-- name: GetAuthTokens :one
SELECT
  consumer_key,
  consumer_secret,
  access_token,
  access_secret,
  access_token_expiry
FROM
  dep_names
WHERE
  name = ?;

-- name: GetAssignerProfile :one
SELECT
  assigner_profile_uuid,
  assigner_profile_uuid_at
FROM
  dep_names
WHERE
  name = ?;

-- name: GetAllDEPNames :many
SELECT name FROM dep_names WHERE tokenpki_staging_cert_pem IS NOT NULL LIMIT ? OFFSET ?;

-- name: GetDEPNames :many
SELECT
  name
FROM
  dep_names
WHERE
  name IN (sqlc.slice('dep_names')) AND
  tokenpki_staging_cert_pem IS NOT NULL
LIMIT ? OFFSET ?;

-- name: StoreAuthTokens :exec
INSERT INTO dep_names
  (name, consumer_key, consumer_secret, access_token, access_secret, access_token_expiry)
VALUES
  (?, ?, ?, ?, ?, ?) AS new
ON DUPLICATE KEY UPDATE
  consumer_key = new.consumer_key,
  consumer_secret = new.consumer_secret,
  access_token = new.access_token,
  access_secret = new.access_secret,
  access_token_expiry = new.access_token_expiry;

-- name: StoreConfig :exec
INSERT INTO dep_names
  (name, config_base_url)
VALUES
  (?, ?) AS new
ON DUPLICATE KEY UPDATE
  config_base_url = new.config_base_url;

-- name: StoreAssignerProfile :exec
INSERT INTO dep_names
  (name, assigner_profile_uuid, assigner_profile_uuid_at)
VALUES
  (?, ?, CURRENT_TIMESTAMP) AS new
ON DUPLICATE KEY UPDATE
  assigner_profile_uuid = new.assigner_profile_uuid,
  assigner_profile_uuid_at = new.assigner_profile_uuid_at;

-- name: StoreCursor :exec
INSERT INTO dep_names
  (name, syncer_cursor)
VALUES
  (?, ?) AS new
ON DUPLICATE KEY UPDATE
  syncer_cursor = new.syncer_cursor;

-- name: StoreTokenPKI :exec
INSERT INTO dep_names
  (name, tokenpki_staging_cert_pem, tokenpki_staging_key_pem)
VALUES
  (?, ?, ?) AS new
ON DUPLICATE KEY UPDATE
  tokenpki_staging_cert_pem = new.tokenpki_staging_cert_pem,
  tokenpki_staging_key_pem = new.tokenpki_staging_key_pem;
