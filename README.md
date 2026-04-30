# Cogito

Project memory MCP server for coding agents While maintaining token efficiency or TYPE SHIT IDK TBH.

Cogito stores durable engineering observations and session summaries per project, then serves that memory back in later sessions and uses two extra skills and a tool at the side such as a caveman talking to reduce yapping of the AI agent and Get Map creating a simple map of the Codebase Cuz yk Map & shit ? Do i need to explain why?.

## What it does

- Creates map yeah thats it tbh
- talk in Direct cold, Precise manner saving token or maybe not idk.(Not gonna claim false stuff)
- Creates project-scoped sessions.
- Stores durable observations (`create_observation`).
- Stores session summaries (`create_summary`) with guard:
  - summary is blocked if no observation exists in that session.
- Fetches context:
  - `get_project_memory` (past sessions, same project)
  - `get_recent_context` (latest observations + summaries)
- Builds codebase substrate map:
  - `get_codebase_map`
- Auto-injects past-session memory into `caveman-review` prompt.

## Install

```bash
git clone https://github.com/DeepanshuChaid/Cogito-Ai.git
cd Cogito-Ai
go install ./cmd/cogito
```

## Quick start

```bash
cogito install
```

`cogito install` writes:

- root `AGENTS.md` policy block
- skills under project
- MCP server registration in `~/.codex/config.toml`

## MCP tools

- `create_observation`
  - Input:
    - `memory` (required)
    - `facts` (optional JSON-array string)
- `create_summary`
  - Input:
    - `request`, `learned`, `nextSteps`
  - Guard:
    - fails if current session has zero observations.
- `get_project_memory`
  - Input:
    - `limit` (optional, default `8`)
  - Returns:
    - past-session observations + summaries for current project.
- `get_recent_context`
  - Input:
    - `limit` (optional, default `10`)
  - Returns:
    - latest observations + summaries for current project.
- `get_codebase_map`
  - Returns:
    - `.cogito/substrate.txt` map.

## Memory model

- DB path: `~/.cogito/cogito.db`
- Core tables:
  - `sdk_sessions`
  - `observations`
  - `session_summaries`
  - `observations_fts` (FTS5)

## CLI commands

- `cogito install`
- `cogito uninstall`
- `cogito serve-mcp`
- `cogito build-map`
- `cogito --help`
- `cogito -v`

## Notes

- Session auto-summary is skipped when no observations were created.
- `get_project_memory` excludes the current active session by design.
