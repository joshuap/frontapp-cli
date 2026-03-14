# 📮 Front CLI

Agent-first CLI for the [Front](https://front.com) API. Every command returns structured HATEOAS JSON that tells the agent what to do next.

## Install

```bash
go install github.com/joshuap/frontapp-cli@latest
```

Or build from source:

```bash
git clone https://github.com/joshuap/frontapp-cli.git
cd frontapp-cli
go build -o front .
cp front ~/.local/bin
```

## Configure

```bash
export FRONT_API_TOKEN="your-api-token"
export FRONT_USER="user@example.com"
```

Or use a config file:

```yaml
# macOS: ~/Library/Application Support/front/config.yaml
# Linux: ~/.config/front/config.yaml
user: user@example.com
```

Environment variables take precedence over config. Run `front config` to verify.

### Token Command

Instead of setting `FRONT_API_TOKEN` directly, you can configure a `token_command` to resolve the token dynamically from a secret manager:

```yaml
token_command:
  - op
  - read
  - op://Vault/front_api_token/password
```

Each list element is a separate argument. The command is executed directly and its stdout is used as the token.

## Usage

```bash
front                                    # list available commands
front config                             # show current configuration
front inboxes                            # list inboxes
front inbox                              # search conversations (default: is:open is:unassigned)
front inbox --assignee user@example.com  # filter by assignee
front read <conversation-id>             # read a conversation thread
```

## Output Format

Every command returns a JSON envelope:

```json
{
  "ok": true,
  "command": "front inboxes",
  "result": {
    "count": 2,
    "inboxes": [
      { "id": "inb_123", "name": "Support" },
      { "id": "inb_456", "name": "Sales" }
    ]
  },
  "next_actions": [
    {
      "command": "front inbox <inbox-id>",
      "description": "Search conversations in this inbox"
    }
  ]
}
```

Errors include a `fix` field:

```json
{
  "ok": false,
  "command": "front inboxes",
  "error": { "message": "no API token provided", "code": "UNAUTHORIZED" },
  "fix": "Set FRONT_API_TOKEN or configure token_command in ~/.config/front/config.yaml"
}
```

## Agent Skill

Install the skill so agents can use the CLI automatically:

```bash
npx skills add joshuap/frontapp-cli
```

Or manually:

```bash
ln -s /path/to/frontapp-cli/skills/front ~/.claude/skills/front
```

## Inspiration

The agent-first HATEOAS design is inspired by [Joel Hooks](https://github.com/joelhooks)' work on [joelclaw](https://github.com/joelhooks/joelclaw) — particularly Joel's [email CLI](https://github.com/joelhooks/joelclaw/blob/main/packages/cli/src/commands/email.ts) and [cli-design skill](https://github.com/joelhooks/joelclaw/blob/main/skills/cli-design/SKILL.md).

## License

MIT
