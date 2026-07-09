# Obsidian MCP Skills

Slash-command workflows for agents using the `obsidian-mcp` server.
Each skill is a `SKILL.md` prompt package that drives the MCP tools
(`read_file`, `write_file`, `append_note`, `list_objects`, `daily_note`,
`search_notes`).

## Skills

### Planned (core)

| Skill | Command | Purpose |
|-------|---------|---------|
| [save-session](save-session/) | `/save-session` | Save a conversation summary note to the vault |
| [daily](daily/) | `/daily` | Read, create, or append to today's daily note |
| [search](search/) | `/search` | Search the vault and load relevant context |

### Potential (assess / refine)

| Skill | Command | Purpose |
|-------|---------|---------|
| [capture](capture/) | `/capture` | Quick-capture a thought to an inbox note |
| [context](context/) | `/context` | Load relevant notes as conversation context |
| [weekly-review](weekly-review/) | `/weekly-review` | Summarize the week's daily notes |
| [moc](moc/) | `/moc` | Generate or update a Map of Content |
| [extract](extract/) | `/extract` | Extract TODOs, questions, or patterns |
| [backlinks](backlinks/) | `/backlinks` | Find notes that link to a given note |
| [template](template/) | `/template` | Create a note from a template |
| [refactor](refactor/) | `/refactor` | Split, merge, or restructure notes |
| [orphans](orphans/) | `/orphans` | Find notes with few or no inbound links |
| [frontmatter](frontmatter/) | `/frontmatter` | Edit or standardize YAML frontmatter |
| [canvas](canvas/) | `/canvas` | Create an Obsidian canvas from notes/context |

## Discovery

Agents discover skills from known roots (`.grok/skills/`, `.claude/skills/`,
`.agents/skills/`, etc.), not from a bare `skills/` directory by default.

To use these skills, either:

1. **Point the agent at this directory** (Grok example in `~/.grok/config.toml`):
   ```toml
   [skills]
   paths = ["~/workspace/obsidian-mcp/skills"]
   ```
2. **Symlink into a discovered root**, e.g.:
   ```bash
   ln -s "$(pwd)/skills" .claude/skills
   # or
   ln -s "$(pwd)/skills" .grok/skills
   ```

## Prerequisites

- `obsidian-mcp` registered as an MCP server
- Vault name matches the `--vault` flag used to start the server
- Write operations need a non-empty `--write-allow` (or paths the skill
  targets must be writable)

## Vault name

Most tools accept a bare path (default vault) or a full URN. Tools that take
an explicit `vault` argument (`daily_note`, `search_notes`) need the logical
vault name from server config. Skills assume the agent knows that name from
MCP setup; if unknown, ask the user once and reuse it.
