# memo

A task stack for your terminal. When you get pulled into something new, push it onto the stack so you don't lose track of what you were doing before.

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

memo list
# → review PR #42 (working for 12m)
#   fix auth bug (paused)

memo switch
# Paused: review PR #42
# Resuming: fix auth bug

memo queue "update docs"
# Queued: update docs

memo pop
# Done: fix auth bug (25m)
# Resuming: review PR #42
```

Running `memo` with no arguments launches an interactive TUI for choosing which task to work on. Use arrow keys to pick a task and press enter to move it to the top of the stack.

```
→ review PR #42 (working for 3m)
  update docs
```

## How it works

A tiny daemon runs in the background, holding your task stack in memory for fast commands. It starts automatically on first use and communicates over a Unix socket at `~/.memo/memo.sock`.

State is persisted to `~/.memo/state.json` on every change, so nothing is lost if the daemon is killed. A log of completed tasks is appended to `~/.memo/log.jsonl`.

## Commands

| Command | Description |
|---|---|
| `memo` | Interactive task picker (falls back to `list` if non-interactive) |
| `memo list` | Show the current task stack |
| `memo push <description>` | Push a new task onto the stack |
| `memo pop` | Complete the current task and resume the previous one |
| `memo switch` | Swap the top two tasks |
| `memo queue <description>` | Add a task to the bottom of the stack |
| `memo --help` | Show help |
