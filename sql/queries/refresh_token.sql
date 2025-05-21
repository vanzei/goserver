-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token, expires_at)
VALUES ($1, $2, NOW() + INTERVAL '60 days')
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT users.* FROM users
JOIN refresh_tokens ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.token = $1
AND revoked_at IS NULL
AND expires_at > NOW();

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens SET revoked_at = NOW(),
updated_at = NOW()
WHERE token = $1
RETURNING *;