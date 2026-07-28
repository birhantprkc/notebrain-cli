# PDF Ingestion with LLMs

NoteBrain supports indexing your PDF attachments using Large Language Models (LLMs) to accurately parse text and structural metadata. Combined with OCR (via Tesseract), this unlocks deep semantic search across all your scanned documents, research papers, and ebooks.

## Requirements

1. **LLM API Key**
2. **`tesseract` (Optional but Recommended)** for Optical Character Recognition of scanned PDFs.

## Supported LLM Providers

NoteBrain automatically detects the provider based on the model prefix or available API keys in your environment.

| Provider           | Environment Variable | Model Syntax Example              |
| :----------------- | :------------------- | :-------------------------------- |
| **OpenRouter**     | `OPENROUTER_API_KEY` | `--llm-model="tencent/hy3"`       |
| **DeepSeek**       | `DEEPSEEK_API_KEY`   | `--llm-model="deepseek-v4-flash"` |
| **OpenAI**         | `OPENAI_API_KEY`     | `--llm-model="gpt-4o"`            |
| **Gemini**         | `GEMINI_API_KEY`     | `--llm-model="gemini-2.5-flash"`  |
| **Ollama** (Local) | `OLLAMA_HOST`        | `--llm-model="ollama/llama3"`     |

## How to Enable

To index PDFs during an ingestion run, supply the `--enable-pdf` and `--llm-model` flags:

```bash
export OPENROUTER_API_KEY="your-key-here"

notebrain ingest --enable-pdf --llm-model="openrouter/inclusionai/ling-3.0-flash:free"
```

You can also set these persistently via the CLI wizard:

```bash
notebrain config init
```

Or in your `~/.notebrain/config.toml`:

```toml
enable_pdf = true
llm_model = "openrouter/inclusionai/ling-3.0-flash:free"
```

## Graceful Fallbacks & Cost Control

If you run `notebrain ingest` with `--enable-pdf` but your API key is missing or unset in the current shell, NoteBrain will gracefully fallback:

- It will print a warning that PDF ingestion is disabled.
- It will **skip** parsing new or updated PDFs.
- It will **preserve** previously ingested PDFs in the ChromaDB index without deleting them.
- It will safely continue ingesting all your standard Markdown (`.md`) files uninterrupted.

This makes it perfectly safe to schedule regular ingestion runs in the background (e.g. via `cron`), without worrying about accidental PDF data loss if your environment variables aren't properly passed.

## Searching PDFs

PDF search results are hidden by default to keep CLI outputs concise. To include PDF contents when searching your vault, use the `--with-pdf` flag:

```bash
notebrain search "machine learning fundamentals" --with-pdf
```
