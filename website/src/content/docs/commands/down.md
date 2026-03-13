---
title: previewctl down
description: Stop a preview environment without deleting its data.
---

## Usage

```bash
previewctl down [name or id]
```

## Description

Stops a preview environment by name or ID.

If no argument is given, PreviewCtl looks for preview environments in the current workspace. If exactly one exists, it is stopped automatically. If multiple exist, they are listed and you must specify one by name or ID.

This command:

1. Looks up the environment by name, falling back to ID
2. Stops and removes all Docker containers on the preview network
3. Removes the Docker network
4. Updates the environment status to `stopped`

Unlike [`delete`](/commands/delete/), `down` **preserves** the environment record, port mappings, and data directory. You can bring the environment back up later with [`previewctl up`](/commands/up/).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `[name or id]` | No | Name or ID of the preview environment to stop. If omitted, auto-selects from the current workspace. |

## Examples

Stop by name:

```bash
previewctl down my-project-main
```

Stop by ID:

```bash
previewctl down abc123
```

Auto-select from current workspace:

```bash
previewctl down
```

Bring it back up later:

```bash
previewctl up
```

## Output

```
stopped frontend
stopped api
stopped postgres
stopped redis
removing network "my-project-main"...
preview environment "my-project-main" stopped
```

## Down vs Delete

| | `down` | `delete` |
|---|--------|----------|
| Stops containers | Yes | Yes |
| Removes network | Yes | Yes |
| Preserves data directory | **Yes** | No |
| Preserves port mappings | **Yes** | No |
| Preserves database records | **Yes** | No |
| Can `up` again | **Yes** | No |
