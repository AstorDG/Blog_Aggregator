-- +goose Up
CREATE TABLE feed_follows(
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
	feed_id UUID NOT NULL UNIQUE references feeds(id) ON DELETE CASCADE
	);

-- +goose Down
DROP TABLE feed_follows
