-- name: GetConfigBaseURL :one
SELECT config_base_url FROM dep_names WHERE name = ?;

-- name: GetSyncerCursor :one
SELECT syncer_cursor FROM dep_names WHERE name = ?;

-- name: GetKeypair :one
SELECT
  tokenpki_cert_pem,
  tokenpki_key_pem
FROM
  dep_names
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
