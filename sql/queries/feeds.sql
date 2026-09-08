-- name: CreateFeed :one
INSERT INTO feeds(id, created_at, updated_at, name, url, user_id)
VALUES (
		$1,
		NOW(),
		NOW(),
		$2,
		$3,
		$4
		)
RETURNING *;

-- name: GetAllFeeds :many
SELECT * FROM feeds;

-- name: GetFeedByURL :one
SELECT * FROM feeds
WHERE $1 = name;

-- name: SetLastFetchedAt :exec
UPDATE feeds
SET feeds.updated_at = NOW() AND feeds.last_fetched_at = NOW()
WHERE $1 = id;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;
