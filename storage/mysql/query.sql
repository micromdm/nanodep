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
