# Fullstack Preview Example

A comprehensive example that exercises **every feature** supported by the previewctl YAML schema.

## Services

| Service | Type | Build | Port | Purpose |
|---------|------|-------|------|---------|
| **postgres** | Pre-built image | `postgres:16` | 5432 | PostgreSQL database |
| **api** | Dockerfile | `node:20-alpine` | 8080 | Express.js REST API |
| **sqlite-api** | Railpack | Python/Flask | 5000 | Lightweight API with SQLite |
| **static-json** | Dockerfile | `nginx:alpine` | 80 | Static JSON file server |
| **frontend** | Nixpacks | React/Vite | 3000 | SPA frontend dashboard |

## Feature Coverage

### Build Types

- **`dockerfile`** — `api`, `static-json` (with custom `dockerfile` path)
- **`nixpacks`** — `frontend`
- **`railpack`** — `sqlite-api`

### Pre-built Images

- **`image`** — `postgres` uses `postgres:16`

### Environment Variables

- **Plain values** — `NODE_ENV: production`, `POSTGRES_DB: appdb`
- **`${Generate(N)}`** — `POSTGRES_PASSWORD`, `JWT_SECRET`, `SESSION_SECRET`, `API_KEY`
- **`${services.X.host}`** — `DATABASE_HOST`, `API_URL`, etc.
- **`${services.X.port}`** — `DATABASE_PORT`, all `*_URL` env vars
- **`${services.X.env.KEY}`** — `DATABASE_URL` reads `postgres.env.POSTGRES_URL`
- **`${secrets.KEY}`** — `STRIPE_API_KEY`, `SENDGRID_API_KEY`
- **`${preview.id}`** — `PREVIEW_ID` on multiple services
- **Bare `${VAR}` self-refs** — `POSTGRES_URL` refs `${POSTGRES_USER}`, `${POSTGRES_PASSWORD}`, `${POSTGRES_DB}`; `APP_LABEL` refs `${APP_NAME}` and `${PREVIEW_ID}`

### Dependency Graph (`depends_on`)

```
postgres ─┐
          ├─> api ──────────┐
sqlite-api ─────────────────├─> frontend
static-json ────────────────┘
```

### Volumes

- `postgres` mounts `/var/lib/postgresql/data`

### Seed Scripts

#### Prestart Seeds (copy-only, no cmd)

- **postgres** — copies custom `postgresql.conf` into `/etc/postgresql/`
- **sqlite-api** — copies pre-populated `seed.db` into `/app/data/`
- **static-json** — copies 3 JSON files into nginx html directory

#### Poststart Seeds (copy + cmd)

- **postgres** — runs SQL migrations then seed data via `psql`
- **sqlite-api** — runs `sqlite3` migration script with `${DB_PATH}` bare env ref in cmd

### Preview Settings

- **`preview.ttl`** — `48h`

## Quick Start

```bash
# 1. Navigate to the example project
cd examples/fullstack

# 2. Copy secrets file and fill in real values
cp .env.example .env

# 3. Initialize the preview environment
previewctl init

# 4. Validate the config
previewctl validate

# 5. Bring up all services
previewctl up
```

## Project Structure

```
fullstack/
├── .previewctl/
│   └── preview.yml          # Full preview config (the star of the show)
├── .env.example              # Required secrets template
├── api/                      # Node.js Express API (dockerfile build)
│   ├── Dockerfile
│   ├── package.json
│   └── server.js
├── frontend/                 # React SPA (nixpacks build)
│   ├── index.html
│   ├── package.json
│   ├── vite.config.js
│   └── src/
│       ├── App.jsx
│       └── main.jsx
├── db/                       # PostgreSQL seed data
│   ├── postgresql.conf       # Prestart: custom config
│   ├── migrations/
│   │   └── 001_create_tables.sql  # Poststart: schema
│   └── seed.sql              # Poststart: sample data
├── sqlite-api/               # Python Flask API (railpack build)
│   ├── app.py
│   ├── requirements.txt
│   └── data/
│       ├── seed.db           # Prestart: initial database
│       └── migrations.sql    # Poststart: run via sqlite3
└── static-json/              # Nginx static server (dockerfile build)
    ├── Dockerfile
    ├── nginx.conf
    └── data/
        ├── config.json       # Prestart: app configuration
        ├── users.json        # Prestart: user fixture data
        └── products.json     # Prestart: product fixture data
```
