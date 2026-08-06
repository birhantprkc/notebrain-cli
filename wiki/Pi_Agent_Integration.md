# Pi Agent Integration

[Pi](https://pi.dev) is an AI coding and terminal agent. You can run Pi in a notebrain-only mode. In this mode, Pi loads one skill: the NoteBrain skill. The agent then works as a semantic retrieval assistant for your Obsidian vault.

The `pinb` shell function sets up this mode. Put the function in `~/.zshrc`.

---

## The Setup (in `~/.zshrc`)

```zsh
# pi — notebrain-only mode: no extensions, skills, templates, or context files
# --session-id pinb: always continues the same notebrain session (per project dir)
pinb() {
  pi -ne -ns -np -nc --skill "$HOME/.agents/skills/notebrain" --session-id pinb "$@"
}
```

---

## Flag Breakdown

Verified against pi v0.83.0 source.

| Flag                                 | Effect                                                                                                                                                                                                   |
| ------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-ne` / `--no-extensions`            | Disables extension discovery (vision proxy, web-access, pi-lens, usage, subagents). An explicit `-e path.ts` still loads.                                                                                |
| `-ns` / `--no-skills`                | Disables skill **discovery**. An explicit `--skill <path>` still loads. In `resource-loader.js`: `noSkills ? mergePaths(cliEnabledSkills, additionalSkillPaths) : …`.                                    |
| `-np` / `--no-prompt-templates`      | No `/explain`, `/brief`, `/commit`, package templates.                                                                                                                                                   |
| `-nc` / `--no-context-files`         | No `AGENTS.md` / `CLAUDE.md` (global and project).                                                                                                                                                       |
| `--skill ~/.agents/skills/notebrain` | Loads only the notebrain skill.                                                                                                                                                                          |
| `--session-id pinb`                  | Reuses the same session on every run. The first run creates it. Later runs load it (`main.js`: `findLocalSessionByExactId` → load, else create). The id must match `^[A-Za-z0-9._-]+$`. `pinb` is valid. |
| `"$@"`                               | Passes through extra arguments to pi. For example: `pinb "find notes about X"`, `pinb -p "…"` (one-shot).                                                                                                |

## Usage

```bash
pinb                                # continue the persistent notebrain session
pinb "what do I know about Redis?"  # launch with a prompt
pinb -p "summarize my notes on K8s" # one-shot, prints result, exits
```
