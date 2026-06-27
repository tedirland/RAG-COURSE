# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A Go-based RAG course project, built incrementally as a learning exercise. A terminal REPL chat client over an OpenAI-compatible LLM API, intended to grow into a retrieval-augmented generation system. Expect the code to be partial and under active construction commit-to-commit.

## Commands

```bash
go run ./cmd/rag      # run the REPL chat app
go build ./...        # build all packages
go test ./...         # run all tests
go test ./chat -run TestFrameAt   # run a single test
```

Requires `OPENAI_API_KEY` in the environment or a `.env` file (loaded via godotenv).

## Configuration

Config is read from environment variables in `config/Load()` (`.env` supported):

- `OPENAI_API_KEY` — API key (required to make real calls)
- `OPENAI_BASE_URL` — defaults to `https://api.openai.com/v1`. Override to point at any OpenAI-compatible endpoint (Ollama, vLLM, LM Studio, etc.)
- `OPENAI_MODEL` — defaults to `gpt-5-mini`
- `SYSTEM_PROMPT_FILE` — path to a file whose contents seed the system prompt
- `DATABASE_URL` — Postgres DSN for the pgvector store. If empty, the store is disabled and the app runs as a plain chat REPL.
- `EMBEDDING_DIM` — embedding vector dimension for the `documents` table; defaults to `768`

## Architecture

Dependencies flow in one direction down a clean chain; each package depends only on the ones below it:

```
cmd/rag/main.go   signal-aware context, wires everything, owns process exit
   -> config            Config struct + Load() from env/.env
   -> app               Run(): builds llm.Client, opens the vector store, hands off to the REPL
   -> chat              REPL loop (RunREPL, Options) over llm.Client.ChatStream
   -> llm               Client wrapping the OpenAI Go SDK; ChatStream() streams deltas, Embed() returns vectors
   -> vector            Document/Result types + Store interface (provider-agnostic)
      -> vector/pgvector  Store implementation backed by Postgres + pgvector
```

Key design points:

- `llm.Message` is the app's own `{role, content}` type. `toSDKMessages()` translates it to the OpenAI SDK's typed message unions (`SystemMessage`/`AssistantMessage`/`UserMessage`), keeping the SDK contained to the `llm` package.
- `ChatStream` streams tokens via an `onDelta func(string)` callback rather than returning the whole response, so the REPL can print as it generates. It accumulates the full message and returns it too.
- `llm.Embedder` is an interface (`Embed(ctx, texts) ([][]float32, error)`) satisfied by `Client`. It batches texts into one `Embeddings.New` call and places each vector by its returned `Index` (the API may reorder results), down-converting `float64` to `float32` for pgvector. Note: it currently uses `cfg.Model` (the chat model) — there is no separate embedding-model config yet.
- The OpenAI SDK is used as a generic LLM interface, not OpenAI-specific — the `BaseURL` override is the seam for swapping providers.
- `vector.Store` is an interface (Upsert/Query/Delete/DeleteBySource/Close); `vector/pgvector` is the first, now-complete implementation. `Query` orders by cosine distance (`<=>`) and returns `Score = 1 - distance`. The interface is the seam for swapping vector backends.
- The vector store is optional. `app.Run` calls `openStore`, which returns `nil` when `DATABASE_URL` is unset; the app then runs as a plain chat REPL with retrieval disabled.

## Current state

`go build ./...` and `go test ./...` both pass. The chat REPL is implemented and wired to the LLM client. The pgvector store is fully implemented (all `Store` methods) and wired into `app.Run`. The `llm.Embedder` client exists but is not yet called anywhere.

Retrieval is still not integrated into the chat loop. The next RAG steps: (1) wire an embedding pipeline to populate the store (chunk + embed + `Upsert`); (2) on each user turn, `Embed` the input, `Query` the store, and inject retrieved context into the prompt. Open follow-ups: add a dedicated embedding-model config (currently shares `OPENAI_MODEL`) and ensure its output dimension matches `EMBEDDING_DIM`.
