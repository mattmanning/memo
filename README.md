# memo

Your brain's task stack. When you get pulled into something new, push it onto the stack so you don't lose track of what you were doing before.

- **Never lose track** of what you were doing before a distraction
- **Instant commands** via in-memory daemon
- **Timestamped log** of all work sessions for later review

## Install

### Homebrew

```
brew tap mattmanning/tap
brew install memo
```

### Go

```
go install github.com/mattmanning/memo@latest
```

### From source

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
# review PR #42 (working for 12m)

memo switch
# Paused: review PR #42
# Resuming: fix auth bug

memo queue "update docs"
# Queued: update docs

memo pop
# Done: fix auth bug (25m)
# Resuming: review PR #42

memo drop
# Dropped: review PR #42 (3m)
# No more tasks.

memo log
# [2026-02-20 14:30] pushed     "fix auth bug" (worked 12m)
# [2026-02-20 14:42] switched   "review PR #42" (worked 3m)
# [2026-02-20 14:55] popped     "fix auth bug" (worked 13m)

memo history
# fix auth bug
#   Started:  2026-02-20 14:30
#   Finished: 2026-02-20 14:55
#   Duration: 25m
```

`memo stack` launches an interactive TUI for choosing which task to work on. Use arrow keys to pick a task and press enter to move it to the top of the stack.

```
memo stack
# → review PR #42 (working for 3m)
#   update docs
```

## How it works

A tiny daemon runs in the background, holding your task stack in memory for fast commands. It starts automatically on first use and communicates over a Unix socket at `~/.memo/memo.sock`.

State is persisted to `~/.memo/state.json` on every change, so nothing is lost if the daemon is killed. A log of completed tasks is appended to `~/.memo/log.jsonl`.

## Commands

| Command | Description |
|---|---|
| `memo` | Show the current task |
| `memo stack` | Interactive task reorder (or show full stack if non-interactive) |
| `memo push <description>` | Push a new task onto the stack |
| `memo pop` | Complete the current task and resume the previous one |
| `memo drop` | Abandon the current task and resume the previous one |
| `memo switch` | Swap the top two tasks |
| `memo queue <description>` | Add a task to the bottom of the stack |
| `memo log` | Show all task activity (pushes, pops, switches) |
| `memo history` | Show completed tasks with start/finish times and durations |
| `memo --help` | Show help |

## Data

All data is stored in `~/.memo/`:

```
~/.memo/
├── memo.sock    # Unix socket for daemon communication
├── memo.pid     # Daemon process ID
├── state.json   # Current task stack
└── log.jsonl    # Timestamped work sessions
```
