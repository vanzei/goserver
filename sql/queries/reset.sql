-- name: ResetDatabase :exec
TRUNCATE TABLE users RESTART IDENTITY CASCADE;