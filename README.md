# memo

A task stack for your terminal. When you get pulled into something new, push it onto the stack so you don't lose track of what you were doing before.

## Install

```
go install github.com/mattmanning/memo@latest
```

Or build from source:

```
git clone https://github.com/mattmanning/memo.git
cd memo
go build -o memo
```

## Usage

```
memo push "fix auth bug"
# Started: fix auth bug

memo push "review PR #42"
# Paused: fix auth bug
# Started: review PR #42

memo
# → review PR #42 (working for 12m)
#   fix auth bug (paused)

memo pop
# Done: review PR #42 (12m)
# Resuming: fix auth bug
```

## How it works

A tiny daemon runs in the background, holding your task stack in memory for fast commands. It starts automatically on first use and communicates over a Unix socket at `~/.memo/memo.sock`.

State is persisted to `~/.memo/state.json` on every change, so nothing is lost if the daemon is killed. A log of completed tasks is appended to `~/.memo/log.jsonl`.

## Commands

| Command | Description |
|---|---|
| `memo` | Show the current task stack |
| `memo push <description>` | Push a new task onto the stack |
| `memo pop` | Complete the current task and resume the previous one |
| `memo --help` | Show help |
