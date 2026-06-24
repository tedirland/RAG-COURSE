# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A Go-based RAG course project, built incrementally as a learning exercise. A terminal REPL chat client over an OpenAI-compatible LLM API, intended to grow into a retrieval-augmented generation system. Expect the code to be partial and under active construction commit-to-commit.

## Commands

```bash
go run ./cmd/rag      # run the REPL chat app
go build ./...        # build all packages
go test ./...         # run all tests
go test ./llm -run TestX   # run a single test
```

Requires `OPENAI_API_KEY` in the environment or a `.env` file (loaded via godotenv).

## Configuration

Config is read from environment variables in `config/Load()` (`.env` supported):

- `OPENAI_API_KEY` — API key (required to make real calls)
- `OPENAI_BASE_URL` — defaults to `https://api.openai.com/v1`. Override to point at any OpenAI-compatible endpoint (Ollama, vLLM, LM Studio, etc.)
- `OPENAI_MODEL` — defaults to `gpt-4o-mini`
- `SYSTEM_PROMPT_FILE` — path to a file whose contents seed the system prompt

## Architecture

Dependencies flow in one direction down a clean chain; each package depends only on the ones below it:

```
cmd/rag/main.go   signal-aware context, wires everything, owns process exit
   -> config      Config struct + Load() from env/.env
   -> app         Run(): constructs llm.Client, hands off to the REPL
   -> llm         Client wrapping the OpenAI Go SDK; ChatStream() streams deltas
   -> chat        REPL loop (RunREPL, Options) — NOT YET IMPLEMENTED
```

Key design points:

- `llm.Message` is the app's own `{role, content}` type. `toSDKMessages()` translates it to the OpenAI SDK's typed message unions (`SystemMessage`/`AssistantMessage`/`UserMessage`), keeping the SDK contained to the `llm` package.
- `ChatStream` streams tokens via an `onDelta func(string)` callback rather than returning the whole response, so the REPL can print as it generates. It accumulates the full message and returns it too.
- The OpenAI SDK is used as a generic LLM interface, not OpenAI-specific — the `BaseURL` override is the seam for swapping providers.

## Current state — build is intentionally broken

`go build ./...` fails: `app/app.go` calls `chat.RunREPL(...)` and `chat.Options{...}`, but the `chat/` package is empty and not imported. The course is mid-step here. The next implementation work is the `chat` package — `RunREPL(ctx, client, Options)` and the `Options` struct (with `SystemPromptFile`) — which drives the read-eval-print loop against `llm.Client.ChatStream`. Do not "fix" this by deleting the call; implement the missing package.
