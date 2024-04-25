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
