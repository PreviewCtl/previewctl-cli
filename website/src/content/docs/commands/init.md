---
title: previewctl init
description: Initialize PreviewCtrl in a repository.
---

## Usage

```bash
previewctl init
```

## Description

Initializes PreviewCtrl in the current directory by creating a `.previewctrl/` folder with a default `preview.yml` configuration file.

Run this once at the root of your project to get started.

## What it creates

```
.previewctrl/
└── preview.yml    # Default config with example services
```

## Example

```bash
cd my-project
previewctl init
```

The generated `preview.yml` includes example services (frontend, API, database, cache) to help you get started. Edit it to match your actual stack.

## Next steps

- Edit `.previewctrl/preview.yml` to define your services
- Run [`previewctl validate`](/commands/validate/) to check your config
- Run [`previewctl up`](/commands/up/) to bring up your environment
