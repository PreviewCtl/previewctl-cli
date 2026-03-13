---
title: Secrets & Environment Variables
description: Provide API keys, passwords, and configuration to your preview services using env files, CLI flags, and generated values.
---

PreviewCtl gives you multiple ways to inject secrets and environment variables into your preview services. This guide covers every source, how they're resolved, and how to use them effectively.

## Sources

There are three ways to provide secrets at deploy time, plus two ways to generate values directly in your config.

### `.env` file

By default, `previewctl up` reads a `.env` file from your project root. Every key-value pair becomes available as a secret.

```ini
# .env
STRIPE_API_KEY=sk_test_abc123
SENDGRID_API_KEY=SG.replace_me
DATABASE_PASSWORD=supersecret
```

```bash
previewctl up
```

Use `--env-file` to load from a different path:

```bash
previewctl up --env-file .env.staging
```

**Format rules:**

- One `KEY=VALUE` per line
- Lines starting with `#` are comments
- `export KEY=VALUE` is supported (the `export` prefix is stripped)
- Quoted values are unquoted: `KEY="hello world"` → `hello world`
- Inline comments are stripped for unquoted values: `KEY=value # comment` → `value`
- If the file doesn't exist, no error is raised — it's treated as empty

:::tip
Create a `.env.example` file in your repo with placeholder values so team members know which secrets are needed:

```ini
# .env.example — copy to .env and fill in real values
STRIPE_API_KEY=sk_test_replace_me
SENDGRID_API_KEY=SG.replace_me
```

:::

### `--secret` flag

Pass secrets directly on the command line. Repeat the flag for multiple values:

```bash
previewctl up --secret STRIPE_API_KEY=sk_test_abc123 --secret DB_PASSWORD=hunter2
```

CLI secrets **override** `.env` file values for the same key.

### `${secrets.<KEY>}` references

In your config, reference provided secrets using the `${secrets.<KEY>}` template:

```yaml
services:
  api:
    build:
      type: dockerfile
      context: ./api
    port: 8080
    env:
      STRIPE_API_KEY: ${secrets.STRIPE_API_KEY}
      SENDGRID_API_KEY: ${secrets.SENDGRID_API_KEY}
    depends_on:
      - postgres
```

PreviewCtl validates that every `${secrets.<KEY>}` reference has a matching value from your `.env` file or `--secret` flags. If a secret is missing, the config fails validation before anything is deployed.

### `${Generate(n)}` — random strings

Generate cryptographically random alphanumeric strings directly in your config. The argument is the string length (1–100):

```yaml
services:
  postgres:
    image: postgres:16
    port: 5432
    env:
      POSTGRES_PASSWORD: ${Generate(16)}
      JWT_SECRET: ${Generate(32)}
```

**Key behavior:**

- Values are generated once and **persisted per preview environment**
- Subsequent `previewctl up` runs for the same environment reuse the same generated values
- This keeps passwords and tokens stable across redeployments

Use `Generate()` for any value that needs to be random but doesn't need to match an external service — database passwords, JWT secrets, session keys, API tokens for internal services.

### `${preview.id}` — preview identifier

Reference the current preview environment's ID:

```yaml
services:
  api:
    build:
      type: dockerfile
      context: ./api
    port: 8080
    env:
      PREVIEW_ID: ${preview.id}
      APP_LABEL: myapp-${preview.id}
```

## Priority order

When the same key exists in multiple sources, later sources win:

| Priority | Source | Description |
|----------|--------|-------------|
| 1 (lowest) | OS environment | Used only for resolving `${secrets.<KEY>}` templates in config |
| 2 | `.env` file | Loaded from project root or `--env-file` path |
| 3 (highest) | `--secret` flag | CLI flags override everything |

Service `env` values in the config can also override injected secrets — if a key appears in both secrets and the service's `env` block, the `env` value takes precedence at the container level.

## Template syntax

Environment variable values support several template expressions:

| Template | Resolves to | Example |
|----------|-------------|---------|
| `${secrets.<KEY>}` | Secret from `.env` or `--secret` | `${secrets.STRIPE_API_KEY}` |
| `${Generate(n)}` | Random alphanumeric string of length n | `${Generate(16)}` |
| `${preview.id}` | Current preview environment ID | `${preview.id}` |
| `${services.<name>.host}` | Hostname of another service | `${services.postgres.host}` |
| `${services.<name>.port}` | Port of another service | `${services.api.port}` |
| `${services.<name>.env.<VAR>}` | Env var from another service | `${services.postgres.env.POSTGRES_URL}` |
| `${<VAR>}` | Self-reference to own env var | `${POSTGRES_USER}` |

### Self-references

A service can reference its own env vars using bare `${VAR}` syntax. This lets you compose values from other entries in the same `env` block:

```yaml
services:
  postgres:
    image: postgres:16
    port: 5432
    env:
      POSTGRES_DB: appdb
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${Generate(16)}
      POSTGRES_URL: postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${services.postgres.host}:${services.postgres.port}/${POSTGRES_DB}?sslmode=disable
```

`POSTGRES_URL` references `POSTGRES_USER`, `POSTGRES_PASSWORD`, and `POSTGRES_DB` — all defined in the same service. PreviewCtl resolves them in dependency order automatically.

### Cross-service references

Reference env vars from other services using `${services.<name>.env.<VAR>}`. The target service must be listed in `depends_on`:

```yaml
services:
  api:
    build:
      type: dockerfile
      context: ./api
    port: 8080
    env:
      DATABASE_URL: ${services.postgres.env.POSTGRES_URL}
      DATABASE_HOST: ${services.postgres.host}
      DATABASE_PORT: ${services.postgres.port}
    depends_on:
      - postgres

  postgres:
    image: postgres:16
    port: 5432
    env:
      POSTGRES_DB: appdb
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${Generate(16)}
      POSTGRES_URL: postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${services.postgres.host}:${services.postgres.port}/${POSTGRES_DB}?sslmode=disable
```

The API service gets the fully resolved `POSTGRES_URL` — including the generated password — without knowing the implementation details.

## Build-time secrets

Secrets from `.env` and `--secret` flags are passed to the build process for all build types (Dockerfile, Nixpacks, Railpack). This means your build steps can access API keys or tokens needed during compilation.

For Dockerfile builds, use `ARG` and `--build-arg` patterns to access secrets during the build:

```dockerfile
ARG NPM_TOKEN
RUN echo "//registry.npmjs.org/:_authtoken=${NPM_TOKEN}" > .npmrc && \
    npm install && \
    rm .npmrc
```

:::caution
Build-time secrets may be cached in image layers. For sensitive values, use multi-stage builds or delete secret files in the same `RUN` step to avoid leaking them into the final image.
:::

## Seed commands

Seed script `cmd` fields support the same template expressions. Use env var references to avoid hardcoding credentials:

```yaml
services:
  postgres:
    image: postgres:16
    port: 5432
    env:
      POSTGRES_DB: appdb
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${Generate(16)}
    seed:
      poststart:
        - source: db/seed.sql
          destination: /tmp/seed.sql
          cmd: psql -U ${POSTGRES_USER} -d ${POSTGRES_DB} -f /tmp/seed.sql
```

All template references in seed commands are validated the same way as service env vars.

## Full example

A complete config using all secret sources together:

```yaml
version: 1

preview:
  ttl: 48h

services:
  postgres:
    image: postgres:16
    port: 5432
    volumes:
      - /var/lib/postgresql/data
    env:
      POSTGRES_DB: appdb
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${Generate(16)}
      POSTGRES_URL: postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${services.postgres.host}:${services.postgres.port}/${POSTGRES_DB}?sslmode=disable
    seed:
      poststart:
        - source: db/seed.sql
          destination: /tmp/seed.sql
          cmd: psql -U ${POSTGRES_USER} -d ${POSTGRES_DB} -f /tmp/seed.sql

  api:
    build:
      type: dockerfile
      context: ./api
    port: 8080
    env:
      NODE_ENV: production
      DATABASE_URL: ${services.postgres.env.POSTGRES_URL}
      JWT_SECRET: ${Generate(32)}
      STRIPE_API_KEY: ${secrets.STRIPE_API_KEY}
      SENDGRID_API_KEY: ${secrets.SENDGRID_API_KEY}
      PREVIEW_ID: ${preview.id}
    depends_on:
      - postgres

  frontend:
    build:
      type: nixpacks
      context: ./frontend
    port: 3000
    env:
      API_URL: http://${services.api.host}:${services.api.port}
    depends_on:
      - api
```

Deploy with:

```bash
# Using .env file (default)
previewctl up

# Or with explicit secrets
previewctl up --secret STRIPE_API_KEY=sk_test_abc --secret SENDGRID_API_KEY=SG.xxx

# Or combine both
previewctl up --env-file .env.staging --secret STRIPE_API_KEY=sk_live_abc
```

## Validation

PreviewCtl catches secret-related errors before deploying:

- **Missing secrets** — `${secrets.X}` referenced but not provided via `.env` or `--secret`
- **Missing service references** — `${services.X.env.Y}` where service `X` doesn't exist or key `Y` isn't defined
- **Missing dependency** — `${services.X.env.Y}` used without `X` in `depends_on`
- **Circular references** — `A` references `B` which references `A`
- **Invalid Generate length** — `${Generate(n)}` where `n` is outside 1–100

Run `previewctl validate` to check your config without deploying:

```bash
previewctl validate --secret STRIPE_API_KEY=test --secret SENDGRID_API_KEY=test
```
