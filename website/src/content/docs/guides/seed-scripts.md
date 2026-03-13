---
title: Seed Scripts
description: Seed containers with files, databases, and migrations using prestart and poststart hooks.
---

PreviewCtl lets you copy files into containers and run initialization commands using **seed scripts**. This is useful for database seeding, running migrations, injecting config files, or loading test fixtures.

## How seeding works

Seeds are defined under `services.<name>.seed` and run at two stages:

**`prestart`** — Runs after the container is created, **before** it starts. Cannot run commands. Use for config files, static data files, and pre-built database files.

**`poststart`** — Runs after the container starts and is **healthy**. Supports `cmd` to execute commands. Use for migrations, SQL seeds, and data imports.

Each seed entry has:

| Field | Required | Description |
|-------|----------|-------------|
| `source` | Yes | Path to the file or directory on the host, relative to the project root. |
| `destination` | Yes | Absolute path inside the container. |
| `cmd` | No | Command to run after copying (poststart only). Executed via `sh -c`. |

## PostgreSQL

Use **poststart** seeds to run migrations and load data after Postgres is ready. The `cmd` field supports environment variable references, so you can reuse credentials defined in `env`.

```yaml
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
    seed:
      prestart:
        - source: db/postgresql.conf
          destination: /etc/postgresql/postgresql.conf
      poststart:
        - source: db/migrations/001_create_tables.sql
          destination: /tmp/001_create_tables.sql
          cmd: >-
            until pg_isready -h 127.0.0.1 -U ${POSTGRES_USER}; do sleep 1; done &&
            psql -h 127.0.0.1 -U ${POSTGRES_USER} -d ${POSTGRES_DB}
            -f /tmp/001_create_tables.sql
        - source: db/seed.sql
          destination: /tmp/seed.sql
          cmd: >-
            until pg_isready -h 127.0.0.1 -U ${POSTGRES_USER}; do sleep 1; done &&
            psql -h 127.0.0.1 -U ${POSTGRES_USER} -d ${POSTGRES_DB}
            -f /tmp/seed.sql
```

**What happens:**

1. `postgresql.conf` is copied into the container before Postgres starts (prestart).
2. After Postgres is running, the migration SQL is copied in and executed.
3. The seed data SQL is copied in and executed next.

:::tip
Use `until pg_isready ...` in your `cmd` to wait for Postgres to accept connections before running SQL.
:::

### Project structure

```
db/
├── postgresql.conf
├── seed.sql
└── migrations/
    └── 001_create_tables.sql
```

**db/migrations/001_create_tables.sql:**

```sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    total DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW()
);
```

**db/seed.sql:**

```sql
INSERT INTO users (name, email) VALUES
    ('Alice Johnson', 'alice@example.com'),
    ('Bob Smith', 'bob@example.com')
ON CONFLICT (email) DO NOTHING;

INSERT INTO orders (user_id, total, status) VALUES
    (1, 99.99, 'completed'),
    (2, 175.00, 'pending');
```

## SQLite

For SQLite, you can either copy a pre-built `.db` file (prestart) or copy a migrations SQL file and run it with a poststart command.

```yaml
services:
  sqlite-api:
    build:
      type: railpack
      context: ./sqlite-api
    port: 5000
    env:
      DB_PATH: /app/data/app.db
      API_KEY: ${Generate(24)}
    seed:
      prestart:
        - source: sqlite-api/data/seed.db
          destination: /app/data/app.db
      poststart:
        - source: sqlite-api/data/migrations.sql
          destination: /tmp/migrations.sql
          cmd: >-
            python3 -c "import sqlite3;
            c=sqlite3.connect('${DB_PATH}');
            c.executescript(open('/tmp/migrations.sql').read());
            c.close()"
```

**What happens:**

1. A pre-built `seed.db` is copied into the container before it starts, giving the app an initial database.
2. After the container is running, `migrations.sql` is copied in and executed via Python's `sqlite3` module.

:::note
The prestart seed copies a complete database file. If you don't have a pre-built `.db` file, you can skip prestart and only use poststart to run SQL migrations against a new database the app creates on startup.
:::

### Project structure

```
sqlite-api/
├── app.py
├── requirements.txt
└── data/
    ├── seed.db
    └── migrations.sql
```

**sqlite-api/data/migrations.sql:**

```sql
CREATE TABLE IF NOT EXISTS notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO notes (id, title, content) VALUES
    (1, 'Welcome', 'This is the first note from the seed data.'),
    (2, 'Setup', 'Preview environment provisioned by previewctl.');
```

## Folder copy (static files)

Use **prestart** seeds to copy entire directories or multiple files into a container before it starts. This is ideal for injecting static assets, configuration files, or data fixtures.

```yaml
services:
  static-json:
    build:
      type: dockerfile
      context: ./static-json
      dockerfile: Dockerfile
    port: 80
    env:
      NGINX_PORT: "80"
    seed:
      prestart:
        - source: static-json/data/config.json
          destination: /usr/share/nginx/html/api/config.json
        - source: static-json/data/users.json
          destination: /usr/share/nginx/html/api/users.json
        - source: static-json/data/products.json
          destination: /usr/share/nginx/html/api/products.json
```

**What happens:**

Each JSON file is copied into the Nginx html directory before the container starts, making them available as static API responses.

### Copying a directory

If `source` points to a directory instead of a file, PreviewCtl copies the entire directory tree recursively:

```yaml
seed:
  prestart:
    - source: static-json/data
      destination: /usr/share/nginx/html/api
```

This copies all files and subdirectories from `static-json/data/` into `/usr/share/nginx/html/api/` inside the container.

### Project structure

```
static-json/
├── Dockerfile
├── nginx.conf
└── data/
    ├── config.json
    ├── users.json
    └── products.json
```

## Combining prestart and poststart

You can mix both stages in a single service. A common pattern is to inject config files before the container starts, then run initialization commands after it's healthy:

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
      prestart:
        - source: db/postgresql.conf
          destination: /etc/postgresql/postgresql.conf
      poststart:
        - source: db/migrations/001_create_tables.sql
          destination: /tmp/001_create_tables.sql
          cmd: >-
            until pg_isready -h 127.0.0.1 -U ${POSTGRES_USER}; do sleep 1; done &&
            psql -h 127.0.0.1 -U ${POSTGRES_USER} -d ${POSTGRES_DB}
            -f /tmp/001_create_tables.sql
```

## Execution order

1. Container is created
2. **Prestart** seeds run in order — files are copied into the stopped container
3. Container starts
4. PreviewCtl waits for the container to be healthy
5. **Poststart** seeds run in order — files are copied, then `cmd` executes (if set)

:::caution
Poststart `cmd` commands run via `sh -c` inside the container. Make sure the required tools (e.g. `psql`, `python3`, `sqlite3`) are available in your container image.
:::
