
-- name: GetConfigBaseURL :one
SELECT config_base_url FROM dep_names WHERE name = $1;

-- name: GetSyncerCursor :one
SELECT syncer_cursor FROM dep_names WHERE name = $1;

-- name: GetCurrentKeypair :one
SELECT
  tokenpki_cert_pem,
  tokenpki_key_pem
FROM
  dep_names
WHERE
  name = $1;

-- name: GetStagingKeypair :one
SELECT
  tokenpki_staging_cert_pem,
  tokenpki_staging_key_pem
FROM
  dep_names
WHERE
  name = $1;

-- name: UpstageKeypair :exec
UPDATE
  dep_names
SET
  tokenpki_cert_pem = tokenpki_staging_cert_pem,
  tokenpki_key_pem = tokenpki_staging_key_pem
WHERE
  name = $1;

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
  name = $1;

-- name: GetAssignerProfile :one
SELECT
  assigner_profile_uuid,
  assigner_profile_uuid_at
FROM
  dep_names
WHERE
  name = $1;


-- name: StoreAuthTokens :exec
INSERT INTO dep_names (
  name, consumer_key, consumer_secret,
  access_token, access_secret,
  access_token_expiry
) VALUES (
  $1, $2, $3, $4, $5, $6
 ) ON conflict (name) DO UPDATE SET
  consumer_key = excluded.consumer_key,
  consumer_secret = excluded.consumer_secret,
  access_token = excluded.access_token,
  access_secret = excluded.access_secret,
  access_token_expiry = excluded.access_token_expiry;


-- name: StoreConfig :exec
INSERT INTO dep_names (
  name, config_base_url
) VALUES ($1, $2) 
ON conflict (name) DO UPDATE SET
config_base_url = excluded.config_base_url;


-- name: StoreAssignerProfile :exec
INSERT INTO dep_names (
  name, assigner_profile_uuid, 
  assigner_profile_uuid_at
) VALUES (
  $1, $2, CURRENT_TIMESTAMP
) ON CONFLICT (name) DO UPDATE SET
assigner_profile_uuid = excluded.assigner_profile_uuid,
assigner_profile_uuid_at = excluded.assigner_profile_uuid_at;

-- name: StoreCursor :exec
INSERT INTO dep_names (
  name, syncer_cursor
) VALUES (
  $1, $2
) ON CONFLICT (name) DO UPDATE SET
syncer_cursor = excluded.syncer_cursor;

-- name: StoreTokenPKI :exec
INSERT INTO dep_names (
  name, tokenpki_staging_cert_pem,
  tokenpki_staging_key_pem
) VALUES (
  $1, $2, $3
) ON CONFLICT (name) DO UPDATE SET
tokenpki_staging_cert_pem = excluded.tokenpki_staging_cert_pem,
tokenpki_staging_key_pem = excluded.tokenpki_staging_key_pem;

-- name: GetAllDEPNames :many
SELECT name FROM dep_names WHERE tokenpki_staging_cert_pem IS NOT NULL LIMIT $1 OFFSET $2;

-- name: GetDEPNames :many
SELECT
  name
FROM
  dep_names
WHERE
  tokenpki_staging_cert_pem IS NOT NULL AND
  name = ANY(sqlc.arg('dep_names')::varchar[])
LIMIT $1 OFFSET $2;
