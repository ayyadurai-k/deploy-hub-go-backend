kear# Setup

Backend for Deploy Hub — Go (Gin + GORM + Postgres).

## Prerequisites

- **Go** `1.26.3` (see `go.mod`)
- **PostgreSQL** running and reachable
- **CompileDaemon** for live reload during development

## 1. Install dependencies

```sh
go mod download
```

## 2. Configure environment

Copy the environment template and fill in the values:

```sh
cp .env.example .env   # if a template exists; otherwise create .env from the keys below
```

`.env` is git-ignored. Required keys:

| Key | Description |
| --- | --- |
| `PORT` | HTTP port (defaults to `8000` if unset) |
| `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` | Postgres connection |
| `DB_SSLMODE`, `DB_TIMEZONE` | Postgres connection options |
| `CORS_ALLOWED_ORIGINS` | Allowed frontend origin(s) |
| `FERNET_KEY` | Key for encrypting OAuth tokens at rest |
| `GOOGLE_OAUTH_CLIENT_ID`, `GOOGLE_OAUTH_CLIENT_SECRET`, `GOOGLE_OAUTH_REDIRECT_URI` | Google OAuth (Google Cloud Console → APIs & Services → Credentials) |
| `GITHUB_OAUTH_CLIENT_ID`, `GITHUB_OAUTH_CLIENT_SECRET`, `GITHUB_OAUTH_REDIRECT_URI` | GitHub OAuth (GitHub → Settings → Developer settings → OAuth Apps) |
| `JWT_SECRET` | Signing secret for session JWTs |
| `OAUTH_STATE_SECRET` | Signing secret for the OAuth `state` parameter |
| `SPA_AUTH_COMPLETE_URL` | Frontend URL to redirect to after auth completes |

## 3. Run

### Development (live reload)

Install CompileDaemon once:

```sh
go install github.com/githubnemo/CompileDaemon@latest
```

Then run — this rebuilds and restarts the server on every file change:

```sh
CompileDaemon -command="./deploy-hub"
```

> `CompileDaemon` watches the source tree, runs `go build` to produce the
> `deploy-hub` binary, then executes `-command` (the freshly built binary).
> Make sure `$(go env GOPATH)/bin` is on your `PATH` so the `CompileDaemon`
> command is found.

### Plain run (no reload)

```sh
go run .
```

### Production build

```sh
go build -o deploy-hub .
./deploy-hub
```

The server listens on `http://localhost:$PORT` (default `8000`). Database
schema is applied automatically via GORM `AutoMigrate` on startup.
