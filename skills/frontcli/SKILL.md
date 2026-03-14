---
name: frontcli
displayName: Front CLI
description: Interact with Front customer support platform via the front CLI. Use when searching conversations, reading messages, or listing inboxes in Front. Triggered by mentions of Front, support tickets, customer conversations, or helpdesk operations.
version: 0.1.0
author: josh
tags:
  - front
  - support
  - email
  - cli
---

# Front CLI

CLI for the Front customer support API. All output is JSON envelopes with `next_actions` for chained workflows.

## Setup

Configure a token command so the CLI can retrieve your API token automatically:

```bash
front config set token_command "op read op://Private/front_api_token/password"
front config set user user@example.com
```

Alternatively, set `FRONT_API_TOKEN` and `FRONT_USER` env vars (these take precedence over config).

Run `front config` to view current configuration.

## Usage

The CLI is self-documenting. Run `front` with no arguments to get available commands and their parameters:

```bash
front
```

Every response includes `next_actions` with the exact commands and parameters to continue. Follow those to navigate.

## Typical Workflows

**Triage unassigned conversations:**

1. `front inboxes` — list available inboxes
2. `front inbox <inbox-id>` — search conversations (defaults to `is:open is:unassigned`)
3. `front read <conversation-id>` — read full conversation thread

**Search with filters:**

- `front inbox --query "is:open is:assigned"` — custom search query
- `front inbox --from user@example.com` — filter by sender
- `front inbox --before 2026-03-01 --after 2026-02-01` — date range
- `front inbox <inbox-id> --limit 10` — scope to inbox with limit

## Rules

- Always start with `front` (no args) if you don't know the available commands — the CLI will tell you.
- Follow `next_actions` from responses rather than guessing commands.
- The `front read` output truncates message text to 500 chars. This is by design for context efficiency.
