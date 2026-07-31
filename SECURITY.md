# Security Policy

## Supported versions

| Version | Supported |
|---|---|
| Latest release | Yes |
| Older releases | Best effort; upgrade first |
| Unreleased `main` | No production support |

## Report a vulnerability

Use GitHub's **Security → Report a vulnerability** private-reporting form for this repository. Do not open a public issue for authentication bypass, data deletion, credential exposure, path traversal, command execution, or container escape findings.

Include the affected version or digest, deployment model, reproduction steps using non-production data, impact, and any suggested mitigation. Remove API tokens, rclone configuration, personal remote names, and media paths from evidence.

Maintainers will acknowledge a complete report as soon as practical, validate impact, prepare a coordinated fix, and credit reporters who request attribution.

## Deployment guidance

- Pin a signed release image by digest.
- Set `ATOMIC_API_TOKEN` and bind to loopback, Tailscale, or an authenticated proxy.
- Keep the source mount read-only until a reviewed move canary.
- Mount only dedicated paths; do not expose the entire host filesystem.
- Protect `rclone.conf` and the SQLite state directory with least privilege.
- Review [the threat model](docs/SECURITY-MODEL.md) before production use.

Never publish real credentials in issues, discussions, CI logs, screenshots, or support requests.
