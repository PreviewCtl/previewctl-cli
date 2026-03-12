---
title: Quick Start
description: Get up and running with PreviewCtl in under a minute.
---

## Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) running locally

## Install

```bash
go install github.com/previewctl/previewctl-cli@latest
```

Verify the installation:

```bash
previewctl --help
```

## Initialize a project

Navigate to your project root and run:

```bash
previewctl init
```

This creates a `.previewctl/` directory with a default `preview.yml` config file.

## Configure your services

Edit `.previewctl/preview.yml` to define your stack:

```yaml
version: 1

preview:
  ttl: 24h

services:
  api:
    build:
      type: dockerfile
      context: .
      dockerfile: Dockerfile
    port: 8080
    env:
      DATABASE_URL: postgresql://postgres:${services.postgres.env.POSTGRES_PASSWORD}@${services.postgres.host}:5432/mydb
    depends_on:
      - postgres

  postgres:
    image: postgres:16
    port: 5432
    env:
      POSTGRES_DB: mydb
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${Generate(16)}
```

## Bring it up

```bash
previewctl up
```

PreviewCtl will:

1. Validate your config
2. Resolve service dependencies (DAG ordering)
3. Build your services (Dockerfile, Nixpacks, or Railpack)
4. Create a Docker network
5. Start all containers in the correct order

## Check your environments

```bash
previewctl list
```

Output:

```
ID        NAME              BRANCH   STATUS   WORKSPACE              CREATED
abc123    my-project-main   main     active   /home/user/my-project  2026-03-11 10:00:00
```

## Clean up

Stop the environment (preserves data so you can `up` again later):

```bash
previewctl down my-project-main
```

Or permanently delete it:

```bash
previewctl delete my-project-main
```

This stops all containers, removes the network, and cleans up the local database records.

## Next steps

- Read the [Configuration reference](/configuration/) for all `preview.yml` options
- Explore individual [commands](/commands/init/) in detail
- Pass secrets with `previewctl up --secret STRIPE_KEY=sk_test_...`
