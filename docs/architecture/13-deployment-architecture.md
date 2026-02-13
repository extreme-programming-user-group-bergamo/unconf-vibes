# 13. Deployment Architecture

## 13.1 Deployment Strategy

**CLI Distribution:**
- GitHub Releases (GoReleaser)
- Homebrew tap for macOS
- Cross-platform: linux, darwin, windows × amd64, arm64

**Server Deployment:**
- Fly.io with Docker
- Tag-triggered via GitHub Actions

## 13.2 CI/CD Pipeline

```yaml
# ci.yaml - Tests on every PR
# release.yaml - GoReleaser on tag
# deploy.yaml - Fly.io deploy on tag
```

## 13.3 Environments

| Environment | API URL | Purpose |
|-------------|---------|---------|
| Development | http://localhost:8080/v1 | Local development |
| Production | https://api.unconf.dev/v1 | Live environment |

## 13.4 Secrets and Key Management

- **Development:** `PASETO_SYMMETRIC_KEY` loaded via local `.env` only; never committed
- **Production:** `PASETO_SYMMETRIC_KEY` managed via Fly.io secrets
- **Rotation Procedure:** deploy with `ACTIVE_KID` + `PREVIOUS_KID` support, rotate every 90 days, retire previous key after max token TTL
- **Break-Glass:** emergency rotation revokes all active sessions and forces re-authentication

## 13.5 Backup and Recovery (SQLite)

- **Backup Cadence:** nightly encrypted snapshot of SQLite volume
- **RPO:** 24 hours
- **RTO:** 4 hours
- **Restore Drill:** run restore rehearsal monthly in non-production environment

---
