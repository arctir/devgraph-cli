# Contributing to Devgraph CLI

## Commit Message Format

This project uses [Conventional Commits](https://www.conventionalcommits.org/) for commit messages. This enables automatic semantic versioning and changelog generation.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type

Must be one of the following:

- **feat**: A new feature (triggers minor version bump)
- **fix**: A bug fix (triggers patch version bump)
- **perf**: A performance improvement (triggers patch version bump)
- **refactor**: Code refactoring (triggers patch version bump)
- **build**: Build system changes (triggers patch version bump)
- **docs**: Documentation only changes (no version bump)
- **style**: Code style changes, formatting (no version bump)
- **test**: Adding or updating tests (no version bump)
- **ci**: CI/CD configuration changes (no version bump)
- **chore**: Other changes that don't modify src or test files (no version bump)
- **revert**: Revert a previous commit (triggers patch version bump)

### Scope (Optional)

The scope should specify the area of the codebase:

- `auth` - Authentication related changes
- `cli` - CLI interface changes
- `config` - Configuration changes
- `entity` - Entity management
- `environment` - Environment management
- `provider` - Provider related changes
- `relation` - Relation management
- `security` - Security related changes
- `workflow` - GitHub Actions workflow changes

### Breaking Changes

To trigger a major version bump, add `BREAKING CHANGE:` in the commit body or footer, or add `!` after the type/scope:

```
feat(auth)!: change authentication flow

BREAKING CHANGE: The authentication flow has been completely rewritten.
Users will need to re-authenticate.
```

### Examples

#### New Feature (Minor Version Bump)
```
feat(entity): add support for batch entity operations

Implement batch create, update, and delete operations for entities
to improve performance when working with multiple entities.
```

#### Bug Fix (Patch Version Bump)
```
fix(auth): correct token refresh logic

Fix issue where tokens were not being refreshed properly when
they expired during a long-running operation.
```

#### Documentation (No Version Bump)
```
docs: update installation instructions

Add instructions for installing on Windows and improve
troubleshooting section.
```

#### Breaking Change (Major Version Bump)
```
feat(api)!: redesign API client interface

BREAKING CHANGE: The API client interface has been redesigned for
better consistency. Users will need to update their code to use
the new interface methods.
```

## Pre-commit Hooks

This project uses pre-commit hooks to enforce commit message format and code quality. To install the hooks:

```bash
pre-commit install
pre-commit install --hook-type commit-msg
```

The hooks will automatically:
- Validate commit messages follow conventional commits format
- Run Go formatting (go-fmt)
- Run Go vet
- Tidy Go modules
- Run unit tests
- Run golangci-lint
- Run gosec security scanner

## Versioning

Version numbers are automatically determined from commit messages:

- `fix:` commits bump the patch version (1.0.0 → 1.0.1)
- `feat:` commits bump the minor version (1.0.0 → 1.1.0)
- Commits with `BREAKING CHANGE:` bump the major version (1.0.0 → 2.0.0)

Releases are created automatically when commits are pushed to the `main` branch.
