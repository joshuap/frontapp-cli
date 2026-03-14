---
name: front
displayName: Front
description: Interact with Front customer support platform via the front CLI. Use when searching conversations, reading messages, or listing inboxes in Front. Triggered by mentions of Front, support tickets, customer conversations, or helpdesk operations.
version: 0.1.0
author: Joshua Wood
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

- `front inbox --assignee user@example.com` — filter by assignee (auto-switches default to `is:assigned`)
- `front inbox --query "is:open is:assigned"` — custom search query
- `front inbox --from user@example.com` — filter by sender
- `front inbox --before 2026-03-01 --after 2026-02-01` — date range
- `front inbox <inbox-id> --limit 10` — scope to inbox with limit

## Search Query Syntax

The `--query` flag accepts Front search syntax. Filters use AND logic between different types, OR for multiple `from:`/`to:`/`cc:`/`bcc:`.

```
is:open|archived|assigned|unassigned|snoozed|trashed|unreplied|waiting|resolved
from:<handle>           sender handle (email, social, etc.)
to:<handle>             recipient (to/cc/bcc)
cc:<handle>             CC recipient
bcc:<handle>            BCC recipient
recipient:<handle>      any role (from, to, cc, bcc)
inbox:<inbox_id>        e.g. inbox:inb_41w25
tag:<tag_id>            e.g. tag:tag_13o8r1
link:<link_id>          linked record
contact:<contact_id>    contact in any recipient role
assignee:<tea_id>       assigned teammate (use alt:email:<email> for email)
participant:<tea_id>    any participating teammate
author:<tea_id>         message author (teammate)
mention:<tea_id>        mentioned teammate
commenter:<tea_id>      commenting teammate
before:<unix_ts>        messages before timestamp (seconds)
after:<unix_ts>         messages after timestamp (seconds)
during:<unix_ts>        messages on same day as timestamp
custom_field:"<name>=<value>"
```

The `--assignee`, `--from`, `--before`, and `--after` flags are shortcuts that append to the query automatically.

## Rules

- Always start with `front` (no args) if you don't know the available commands — the CLI will tell you.
- Follow `next_actions` from responses rather than guessing commands.
- The `front read` output truncates message text to 500 chars. This is by design for context efficiency.
