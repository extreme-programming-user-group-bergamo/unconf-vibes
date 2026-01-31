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

---
