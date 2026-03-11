---
title: Configuration
description: Full reference for .previewctrl/preview.yml.
---

PreviewCtl is configured via a `.previewctrl/preview.yml` file at the root of your project. This page documents every available option.

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
