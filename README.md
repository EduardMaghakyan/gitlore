# gitlore

When coding agents generate code, the *why* behind a change evaporates at commit time. Git records what changed, never what was intended.

gitlore fixes this. It captures agent conversations at commit time, distills them into a short summary, and attaches it as a Git note — automatically, without changing how you work.

## What you get

Every commit made with an AI coding agent gets a note like this:

```
$ git log

commit a1b2c3d
Author: you
Date:   Wed Mar 26 10:30:00 2026

    Fix token refresh race condition

Notes:
    [agent-assisted] The developer was trying to fix a race condition in the
    OAuth token refresh flow. The auth middleware now queues concurrent refresh
    requests instead of firing them in parallel. Error handling for expired
    refresh tokens is stubbed but not yet implemented.
```

Three sentences: what was intended, what changed, what's left. Visible in `git log`, `git show`, and any Git UI that displays notes.

## Install

```bash
go install github.com/eduardmaghakyan/gitlore@latest
```

Then, in any repo:

```bash
gitlore install
```

That's it. Keep using `git` normally. Notes appear automatically after each commit that involved an AI agent conversation.

## How it works

1. You commit as usual (`git commit`)
2. A post-commit hook fires in the background — your commit returns instantly
3. gitlore reads the agent conversation since your last commit
4. It distills the conversation + diff into a 3-sentence summary
5. The summary is attached as a Git note

No new habits. No new commands. No blocking.

## Commands

| Command | What it does |
|---|---|
| `gitlore install` | Set up hooks in the current repo |
| `gitlore amend-note [SHA]` | Edit a note in your `$EDITOR` |
| `gitlore show [SHA]` | Show the note on a commit |
| `gitlore log` | Git log with notes |
| `gitlore push` | Push commits + notes to remote |

## Configuration

Optional. Create `~/.gitlore` (global) or `.gitlore` (per-repo):

```toml
[distill]
use_cli = true              # use local claude CLI (default)
model = "claude-sonnet-4-6" # model for summarization

[notes]
auto_push = false           # push notes on every gitlore push
```

Distillation uses the local `claude` CLI by default. If unavailable, falls back to the Anthropic API using `ANTHROPIC_API_KEY`.

## Principles

- **Zero new habits.** Works with existing `git` commands.
- **Non-blocking.** Never slows down or prevents a commit.
- **Transparent.** You always see what was captured. Edit any note with `gitlore amend-note`.
- **Local-first.** No accounts, no cloud dependency. Notes live in the repo.

## License

MIT
