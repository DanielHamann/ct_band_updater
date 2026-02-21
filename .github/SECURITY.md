# Security Policy

## Supported Versions

Only the latest release receives security fixes.

| Version | Supported |
|---------|-----------|
| Latest  | ✅        |
| Older   | ❌        |

## Reporting a Vulnerability

Please **do not** open a public GitHub issue for security vulnerabilities.

Instead, report them via [GitHub's private vulnerability reporting](https://github.com/DanielHamann/ct_band_updater/security/advisories/new).

We will respond as quickly as possible and coordinate a fix before any public disclosure.

## Security Notes

- The ChurchTools API key is stored exclusively in the OS keychain (macOS Keychain / Windows Credential Manager) — never in plain text on disk.
- All local app data is written to `~/.ct_musikteam/data.json` with owner-only permissions (`0600`).
- The app communicates only with the ChurchTools instance URL you configure — no other external services.
