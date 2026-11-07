-- +goose Up
--id - a unique identifier for the post
--created_at - the time the record was created
--updated_at - the time the record was last updated
--title - the title of the post
--url - the URL of the post (this should be unique)
--description - the description of the post
--published_at - the time the post was published
--feed_id - the ID of the feed that the post came from
--additionally, we want to ensure that we don't have duplicate posts for the same feed. Therefore, we will add a unique constraint on the combination of feed_id and url.
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    description TEXT,
    published_at TIMESTAMP NOT NULL,
    feed_id UUID NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
    UNIQUE(feed_id, url)
);

-- +goose Down
DROP TABLE posts;