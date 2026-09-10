# migrations

SQL migrations for this service, applied with [golang-migrate](https://github.com/golang-migrate/migrate).

## Convention

Each migration is a **pair** of files:

```
000001_create_users_table.up.sql     -- applies the change
000001_create_users_table.down.sql   -- reverts it
```

- Version is a zero-padded sequence: `000001`, `000002`, ...
- `up` runs on `migrate up`; `down` runs on `migrate down`.

## Commands

Create a new pair:

```bash
migrate create -ext sql -dir migrations -seq <name>
```

Apply all pending migrations:

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

Check the current version:

```bash
migrate -path migrations -database "$DATABASE_URL" version
```

Delete this file once the service has its first real migration.
