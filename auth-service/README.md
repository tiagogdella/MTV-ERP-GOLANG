# auth-service

Authentication service for the MTV ERP. It manages users and issues JSON Web
Tokens used by the other services to authorize requests.

## What it does

- Stores users in PostgreSQL (via GORM), with passwords hashed using bcrypt.
- Issues signed JWTs on login (HS256, 24h expiry).
- Validates tokens on behalf of other services — they never verify a token
  themselves, they call `ValidateToken`.

Exposed over gRPC (package `auth`, service `AuthService`):

| RPC             | Request                    | Response                      |
|-----------------|----------------------------|-------------------------------|
| `CreateUser`    | `email, password, role`    | `user_id`                     |
| `Login`         | `email, password`          | `token`                       |
| `ValidateToken` | `token`                    | `valid, user_id, role`        |

The JWT payload carries `sub` (user id), `role`, `iat` and `exp`.

## Ports

| Port                  | Protocol | Purpose                          |
|-----------------------|----------|----------------------------------|
| `PORT` (default 8080) | HTTP     | `/healthz`, `/readyz`, `/metrics`|
| `GRPC_PORT` (9090)    | gRPC     | `AuthService`                    |

## Environment variables

| Variable       | Required | Default | Notes                                             |
|----------------|----------|---------|---------------------------------------------------|
| `DATABASE_URL` | yes      | —       | Postgres DSN, e.g. `postgres://user:pass@host:5432/auth?sslmode=disable` |
| `JWT_SECRET`   | yes      | —       | HMAC secret used to sign and verify tokens        |
| `PORT`         | no       | `8080`  | HTTP port                                         |
| `GRPC_PORT`    | no       | `9090`  | gRPC port                                         |
| `ENVIRONMENT`  | no       | `local` | Free-form label, shown in logs                    |

The process exits on startup if `DATABASE_URL` or `JWT_SECRET` is missing.

## Running locally

Start a Postgres and run the migrations:

```bash
docker run -d --name auth-db \
  -e POSTGRES_DB=auth -e POSTGRES_USER=auth_service -e POSTGRES_PASSWORD=secret \
  -p 5432:5432 postgres:16

export DATABASE_URL='postgres://auth_service:secret@localhost:5432/auth?sslmode=disable'

migrate -path migrations -database "$DATABASE_URL" up
```

Then start the service:

```bash
export JWT_SECRET='dev-secret'
make run
```

Smoke test:

```bash
curl -s localhost:8080/healthz            # -> ok
grpcurl -plaintext -import-path proto -proto auth.proto \
  -d '{"email":"a@b.com","password":"123456","role":"operador"}' \
  localhost:9090 auth.AuthService/CreateUser
```

## Make targets

| Target       | Action                          |
|--------------|---------------------------------|
| `make build` | build the binary into `bin/`    |
| `make run`   | run the server                  |
| `make test`  | run all tests                   |
| `make lint`  | `go vet ./...`                  |

## Tests

- Unit tests: `internal/auth` (bcrypt, JWT round-trip).
- Integration test: `internal/grpcserver` spins up a real Postgres with
  Testcontainers, runs the migrations, and exercises `Login` end to end.
  **Docker must be running.**

## Protobuf

The `.proto` lives in `proto/`. Regenerate the Go code (`internal/authpb`) with:

```bash
cd proto && buf generate
```

## Deploy

Kubernetes manifests are in `../deploy/auth-service/`. The manual deploy flow to
the company k3s cluster is documented in `../docs/deploy.md`.
