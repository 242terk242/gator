# Gator

Gator is a command-line RSS feed aggregator written in Go. It stores users, feeds, subscriptions, and posts in PostgreSQL, and lets you manage feed subscriptions and periodically collect feed entries.

## Features

- Register users and switch between them.
- Add RSS feeds and follow or unfollow feeds.
- List feeds and the feeds followed by the current user.
- Periodically fetch feed entries and store them in PostgreSQL.
- Browse stored posts.

## Technologies

- Go 1.26
- PostgreSQL
- [`sqlc`](https://sqlc.dev/) for generated database access code
- [`lib/pq`](https://github.com/lib/pq) PostgreSQL driver
- [`google/uuid`](https://github.com/google/uuid) for UUIDs

## Installation

### Prerequisites

- Go 1.26 or later
- PostgreSQL
- `psql`, to initialize the database

### Set up

1. Clone the repository and enter the project directory:

   ```sh
   git clone https://github.com/242terk242/gator.git
   cd gator
   ```

2. Create a PostgreSQL database. For example:

   ```sh
   createdb gator
   ```

3. Apply the schema files in order. Set `DATABASE_URL` to your PostgreSQL connection string before running these commands:

   ```sh
   export DATABASE_URL="postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"

   psql "$DATABASE_URL" -f sql/schema/001_users.sql
   psql "$DATABASE_URL" -f sql/schema/002_feeds.sql
   psql "$DATABASE_URL" -f sql/schema/003_feed_follows.sql
   psql "$DATABASE_URL" -f sql/schema/004_last_fetched_at.sql
   psql "$DATABASE_URL" -f sql/schema/005_posts.sql
   ```

   Replace the example connection string with credentials and a database name appropriate for your setup.

4. Create `~/.gatorconfig.json`. Gator reads the database URL and current username from this file:

   ```json
   {
     "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
     "current_user_name": ""
   }
   ```

   The `db_url` value must match your PostgreSQL connection string. The application reads this setting from the config file; exporting `DATABASE_URL` alone does not configure Gator.

5. Install the CLI:

   ```sh
   go install github.com/242terk242/gator@latest
   ```

   This installs the `gator` executable into Go's binary directory (usually `$(go env GOPATH)/bin`). Make sure that directory is on your `PATH`. This command is available once the repository is published and accessible at that module path. To build from a local checkout instead, run `go build -o gator .` and invoke it as `./gator`.

## Usage

Run the installed CLI as `gator <command> [arguments]`. If you built it from a local checkout, use `./gator` instead.

Register a user and make that user current:

```sh
gator register alice
```

Registering also sets the current user. To switch to an existing user later, use `login`:

```sh
gator login alice
gator users
```

Add a feed (the current user automatically follows it), or follow an existing feed by URL:

```sh
gator addfeed "Wagslane Blog" "https://www.wagslane.dev/index.xml"
gator follow "https://example.com/feed.xml"
gator following
gator feeds
```

Start periodic aggregation with a Go duration such as `30s`, `1m`, or `1h`. Stop it with `Ctrl+C`. Browse the posts stored in the database with:

```sh
gator agg 1m
gator browse
```

To stop following a feed, pass its URL:

```sh
gator unfollow "https://example.com/feed.xml"
```

### Commands

| Command | Description |
| --- | --- |
| `register <name>` | Create a user and set them as current. |
| `login <name>` | Set the current user to an existing user. |
| `users` | List registered users and identify the current user. |
| `addfeed <name> <url>` | Add a feed and follow it as the current user. |
| `feeds` | List feeds. |
| `follow <url>` | Follow an existing feed as the current user. |
| `following` | List feeds followed by the current user. |
| `unfollow <url>` | Stop following a feed as the current user. |
| `agg <duration>` | Fetch feeds periodically using the given Go duration. |
| `browse` | List stored posts. |
| `reset` | Delete all users. **This is destructive**; related records may also be removed through database cascades. |

## Contributing

Contributions are welcome. To get started:

1. Fork the repository and create a branch for your change.
2. Keep changes focused, and include or update tests when appropriate.
3. Format changed Go files with `gofmt`, then run the project's checks:

   ```sh
   go test ./...
   go build ./...
   ```

4. Open a pull request describing the change and any relevant testing.

For database changes, keep the SQL schema and query files in `sql/schema/` and `sql/queries/` consistent with the generated code in `internal/database/`.

## License

No license file or license declaration is currently present in this repository. Check with the maintainers before using, modifying, or distributing this project.
