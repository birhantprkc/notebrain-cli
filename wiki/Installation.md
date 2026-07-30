# Installation

## Prerequisites

- Go 1.26.4 or higher.
- `make` (optional, to build from source).
- A CGO-enabled toolchain (GCC or Clang on Linux or macOS). The embedded vector database requires C and C++ bindings.

## Install Pre-built Binaries (Recommended)

1. Download the pre-built Linux or macOS binaries from the [GitHub Releases](https://github.com/nmdra/notebrain-cli/releases) page.
2. Extract the archive.
3. Put the `notebrain` binary in your `$PATH`.

## Build from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/nmdra/notebrain-cli.git
   cd notebrain-cli
   ```

2. Use Make to build the binary:
   ```bash
   make build
   ```
   Note: This command executes `CGO_ENABLED=1 go build -o notebrain .`

3. Move the binary to a directory in your `$PATH`:
   ```bash
   sudo mv notebrain /usr/local/bin/
   ```

## Configuration

NoteBrain uses a TOML file for persistent configuration. The default location is `~/.notebrain/config/config.toml`.

1. Create the configuration directory:
   ```bash
   mkdir -p ~/.notebrain/config
   ```

2. Copy the template from the repository:
   ```bash
   cp config.example.toml ~/.notebrain/config/config.toml
   ```

3. Edit `~/.notebrain/config/config.toml`. Set your vault path, the database storage location, and the default format:
   ```toml
   vault-path = "/path/to/your/Obsidian Vault"
   vault-name = "My Vault"
   chroma-path = "~/.notebrain/chroma"
   format = "text"
   ```

By default, NoteBrain stores the local ChromaDB database at `~/.notebrain/chroma`. You can override any setting with command-line flags (for example, `--chroma-path`, `--vault-path`, `--format`).
