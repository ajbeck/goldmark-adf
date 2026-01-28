# Git Workflows

Safe by default: `git status/diff/log`. Push only when user asks.

## Branch Operations

- `git checkout` ok for PR review / explicit request
- Branch changes require user consent
- Destructive ops forbidden unless explicit (`reset --hard`, `clean`, `restore`, `rm`, …)
- Don't delete/rename unexpected stuff; stop + ask
- No repo-wide S/R scripts; keep edits small/reviewable

## Stashing

- Avoid manual `git stash`; if Git auto-stashes during pull/rebase, that's fine

## User Consent

- If user types a command ("pull and push"), that's consent for that command
- No amend unless asked

## Code Review

- Big review: `git --no-pager diff --color=never`
- Multi-agent: check `git status/diff` before edits; ship small commits

## Commit Messages

Use the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification.

Format: `<type>[optional scope]: <description>`

Types:

- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation only
- `style` - Formatting, no code change
- `refactor` - Code change that neither fixes a bug nor adds a feature
- `test` - Adding or correcting tests
- `chore` - Maintenance tasks
- `ci` - Continuous Integration changes

Examples:

```
feat: add terraform MCP interface
fix: resolve SSL certificate issue for corporate proxy
docs: add contributing guide to README
chore: bump version to 0.7.0
```

## Committing Changes

When the user asks to commit, use the `committer` script:

```sh
# If agent-scripts/bin is in PATH
committer "feat: description of change" file1.ts file2.ts

# Or use full path
./agent-scripts/bin/committer "feat: description of change" file1.ts file2.ts
```

The script handles staging and committing. Use conventional commit format for the message.

Do NOT manually run `git add` and `git commit` - always use `committer` instead.
