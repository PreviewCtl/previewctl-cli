---
title: Configuration
description: Full reference for .previewctl/preview.yml.
---

PreviewCtl is configured via a `.previewctl/preview.yml` file at the root of your project. This page documents every available option.

## Full example

```yaml
version: 1

preview:
  ttl: 24h

services:
  frontend:
    build:
      type: dockerfile
      context: .
      dockerfile: frontend/Dockerfile
    port: 3000
    env:
      API_URL: http://${services.api.host}:${services.api.port}
    depends_on:
      - api

  api:
    build:
      type: railpack
      context: ./api
    port: 8080
    env:
      DATABASE_HOST: ${services.postgres.host}
      DATABASE_PORT: ${services.postgres.port}
      REDIS_HOST: ${services.redis.host}
      REDIS_PORT: ${services.redis.port}
      DATABASE_URL: ${services.postgres.env.POSTGRES_URL}
      STRIPE_API_KEY: ${secrets.STRIPE_API_KEY}
    depends_on:
      - postgres
      - redis

  worker:
    build:
      type: nixpacks
      context: ./worker
    env:
      REDIS_HOST: ${services.redis.host}
      REDIS_PORT: ${services.redis.port}
    depends_on:
      - redis

  postgres:
    image: postgres:16
    port: 5432
    volumes:
      - /var/lib/postgresql/data
    seed:
      poststart:
        - source: db/seed.sql
          destination: /tmp/seed.sql
          cmd: psql -U ${POSTGRES_USER} -d ${POSTGRES_DB} -f /tmp/seed.sql
    env:
      POSTGRES_DB: mydb
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${Generate(16)}
      POSTGRES_URL: postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${services.postgres.host}:${services.postgres.port}/${POSTGRES_DB}?sslmode=disable

  redis:
    image: redis:7
    port: 6379
```

## Top-level fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | `int` | Yes | Config version. Currently `1`. |
| `preview` | `object` | No | Global preview settings. |
| `services` | `map` | Yes | Map of service name to service config. |

## `preview`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ttl` | `string` | — | Time-to-live for the preview environment (e.g. `24h`, `30m`). Environment is cleaned up after this duration. |

## `services.<name>`

Each key under `services` defines a service. A service is either **built from source** (has `build`) or **uses a pre-built image** (has `image`).

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `build` | `object` | No* | Build configuration. |
| `image` | `string` | No* | Docker image to use (e.g. `postgres:16`). |
| `port` | `int` | No | Port the service listens on inside the container. |
| `env` | `map` | No | Environment variables. Values support template syntax. |
| `volumes` | `string[]` | No | Container volume mounts. |
| `seed` | `object` | No | Seed configuration for initializing the container with files and commands. |
| `depends_on` | `string[]` | No | Services this service depends on. Controls startup order. |

\* One of `build` or `image` is required.

## `services.<name>.build`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | `string` | Yes | Build type: `dockerfile`, `nixpacks`, or `railpack`. |
| `context` | `string` | Yes | Build context path, relative to the project root. |
| `dockerfile` | `string` | No | Path to Dockerfile (only for `type: dockerfile`). Relative to context. |

### Build types

| Type | Description |
|------|-------------|
| `dockerfile` | Standard Docker build. Specify `context` and optionally `dockerfile`. |
| `nixpacks` | Automatic build using [Nixpacks](https://nixpacks.com). Detects language and framework. |
| `railpack` | Automatic build using [Railpack](https://railpack.com). |

## `services.<name>.seed`

The `seed` field lets you copy files into a container and optionally run initialization commands. This is useful for database seeding, loading fixtures, or running migrations.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `prestart` | `SeedEntry[]` | No | Files to copy into the container **before** it starts. No commands can be run. |
| `poststart` | `SeedEntry[]` | No | Files to copy and commands to run **after** the container starts and is healthy. |

### Seed entry fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `source` | `string` | Yes | Path to the file or directory on the host, relative to the project root. |
| `destination` | `string` | Yes | Path inside the container where the file will be copied. |
| `cmd` | `string` | No | Command to run inside the container after copying (only for `poststart`). |

### Example: Database seeding

```yaml
services:
  postgres:
    image: postgres:16
    port: 5432
    seed:
      poststart:
        - source: db/seed.sql
          destination: /tmp/seed.sql
          cmd: psql -U ${POSTGRES_USER} -d ${POSTGRES_DB} -f /tmp/seed.sql
    env:
      POSTGRES_DB: mydb
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${Generate(16)}
```

### Example: Config file injection

```yaml
services:
  nginx:
    image: nginx:latest
    port: 80
    seed:
      prestart:
        - source: config/nginx.conf
          destination: /etc/nginx/nginx.conf
```

### How seeding works

1. **Prestart seeds** run after the container is created but before it starts. Files are copied into the stopped container. This is ideal for configuration files that must be in place before the process launches.
2. **Poststart seeds** run after the container starts and is healthy. Files are copied first, then the optional `cmd` is executed inside the container via `sh -c`. This is ideal for database initialization, running migrations, or loading fixtures.

## Template syntax

Environment variable values support template expressions:

| Template | Description | Example |
|----------|-------------|---------|
| `${services.<name>.host}` | Hostname of another service | `${services.postgres.host}` |
| `${services.<name>.port}` | Port of another service | `${services.api.port}` |
| `${services.<name>.env.<VAR>}` | Env var from another service | `${services.postgres.env.POSTGRES_URL}` |
| `${secrets.<KEY>}` | Injected secret | `${secrets.STRIPE_API_KEY}` |
| `${Generate(n)}` | Random string of length `n` | `${Generate(16)}` |
| `${<VAR>}` | Self-reference to own env var | `${POSTGRES_USER}` |

### Secrets

Secrets referenced with `${secrets.<KEY>}` are provided at runtime via:

```bash
# Inline
previewctl up --secret STRIPE_API_KEY=sk_test_abc123

# From file
previewctl up --env-file .env
```
