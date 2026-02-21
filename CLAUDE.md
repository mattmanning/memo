# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build -o memo          # build binary
go vet ./...              # lint
```

There are no tests yet. The project has zero external dependencies (stdlib only).

## Architecture

memo is a task stack CLI tool with a client-daemon architecture. All files are in `package main` with no subdirectories.

- **main.go** — CLI entry point, parses commands (`push`, `pop`, `list`, `__daemon`, `--help`) and dispatches to client functions
- **daemon.go** — Background HTTP server listening on a Unix socket (`~/.memo/memo.sock`). Holds the `TaskStack` in memory, protected by a mutex. Exposes `GET /stack`, `POST /push`, `POST /pop`. The daemon is auto-started by `ensureDaemon()` which spawns `memo __daemon` as a detached process
- **client.go** — HTTP client that connects to the daemon over the Unix socket and formats output for the terminal
- **stack.go** — `TaskStack` and `Task` data structures (slice-based stack, newest task at index 0)
- **persist.go** — Atomic state save/load (`~/.memo/state.json`) and append-only task log (`~/.memo/log.jsonl`)

The daemon persists state via `SaveState` (atomic write with tmp+rename) on every mutation and logs completed/paused tasks to `log.jsonl` via `LogTaskStop`.

## Conventions

- Do not add a `Co-Authored-By` line to commit messages.
