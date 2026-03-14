# Front CLI

Agent-first CLI for the Front API. HATEOAS JSON envelope output with `next_actions`.

## Build & Test

```
go build -o front .
go test ./...
go vet ./...
```

## Commits

Use conventional commits: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`.

## Architecture

- `cmd/` — Cobra commands. One file per command. Types in `types.go`.
- `internal/api/` — Generated OpenAPI client. Do not edit `front.gen.go`.
- `internal/config/` — Config file loading, token resolution.
- `internal/envelope/` — JSON envelope output format.
- `skills/front/` — Agent skill definition.

## Key Patterns

- All output goes through `envelope.FprintResult` / `envelope.FprintError`.
- Token resolution: `FRONT_API_TOKEN` env > `token_command` from config > error.
- User resolution: `FRONT_USER` env > `user` from config > empty.
- Teammate references use `alt:email:<email>` format.
