# Security Policy

## Supported versions

| Version | Supported |
|---|---|
| Latest release | Yes |
| Older releases | Best effort; upgrade first |
| Unreleased `main` | No production support |

## Report a vulnerability

Use GitHub's **Security → Report a vulnerability** private-reporting form for this repository. Do not open a public issue for authentication bypass, unintended data deletion, credential exposure, path traversal, command execution, or container escape findings.

Include the affected version or image digest, deployment model, minimal reproduction steps using non-production data, impact, and any suggested mitigation. Remove API tokens, rclone configuration, personal remote names, and media paths from all evidence.

Maintainers will acknowledge a complete report as soon as practical, validate impact, prepare a coordinated fix, and credit reporters who request attribution.

## Deployment guidance

- Pin a signed release image by immutable digest and verify its Cosign identity, SBOM, and provenance.
- Set a random `ATOMIC_API_TOKEN` of at least 32 characters and bind the UI to loopback, Tailscale, or an authenticated proxy.
- Keep the reference source bind read-only for copy deployments and all dry-runs.
- Treat real `move` mode as destructive. Require `deleteSource=true`; use `fail` for a clean destination or `merge-immutable` only after reviewing a partial unit, prefer `checksum` for stronger evidence on transferred paths, and never manually clean source overlaps without independent content proof because `--ignore-existing` skips their comparison. Make only the narrow required source subtree writable, stop every writer, confirm the exact job, and begin with one stable canary unit.
- Keep the default 30-day stable window in production. A three-day window is only a scoped canary setting.
- Mount only dedicated source, state, and rclone-config paths; never expose the complete host filesystem or the host-global rclone configuration.
- Protect `atomic-sync.db` and the dedicated rclone config directory with least privilege. Only the private config directory needs write access for OAuth token refresh.
- Configure a dedicated Google OAuth client before increasing Drive throughput. Do not publish the client secret or refreshed tokens.
- Remember that rclone performs direct object transfers. A move is not a cross-provider ACID directory transaction, and an interrupted operation can require branch analysis and a controlled retry.
- Treat pre-v0.2 `.atomic-sync-staging` as separate recovery evidence. Version 0.2 neither manages nor deletes it; inventory and independently verify it before explicit manual cleanup.
- Review [the full threat model](docs/SECURITY-MODEL.md) and [production runbook](docs/OPERATIONS.md) before enabling writes.

Never publish real credentials in issues, discussions, CI logs, screenshots, or support requests.
