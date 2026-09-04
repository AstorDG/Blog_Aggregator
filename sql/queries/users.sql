-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, name)
VALUES (
		$1,
		$2,
		$3,
		$4
		)
RETURNING *;

-- name: GetUserByName :one
SELECT * from users
WHERE $1 = name;

-- name: ResetDatabase :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT name from users;

-- name: GetUserNameById :one
SELECT name from users
WHERE $1 = id;
