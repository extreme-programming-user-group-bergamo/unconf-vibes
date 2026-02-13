# Next Steps

1. **Epic 1 Stories:** Begin with project scaffolding (Story 1.1)
2. **CI/CD Setup:** Configure GitHub Actions early
3. **Development Environment:** Docker Compose + MailHog
4. **First Milestone:** Working `unconf login` command

## Dependency Health Check (2026-02-14)

- **Current status:** No archived dependencies detected in the core architecture stack.
- **Keep as-is:** Cobra, Viper, Bubble Tea, Lip Gloss, Gin, go-resty, validator, testify, go-keyring, golang-migrate, wneessen/go-mail.
- **Watch item:** `mattn/go-sqlite3` requires CGO; if cross-compilation friction increases, evaluate `modernc.org/sqlite`.
- **Watch item:** Viper remains widely used but heavy for strict typed config; evaluate `knadh/koanf` if config complexity grows.
- **Watch item:** Track `aidanwoods.dev/go-paseto` module updates quarterly (latest observed: `v1.6.0`).
- **Recheck trigger:** Run dependency audit before each minor release or every 90 days.

---
