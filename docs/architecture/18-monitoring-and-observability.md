# 18. Monitoring and Observability

## 18.1 Monitoring Stack

| Component | Tool |
|-----------|------|
| Logging | log/slog (JSON handler) |
| Log Aggregation | Fly.io logs |
| Metrics | Fly.io dashboard |
| Health Check | `/health` endpoint |

## 18.2 Key Metrics

- Request count by endpoint
- Response time percentiles
- Error rate
- Bookings created/cancelled
- Email delivery success rate
- Auth success/failure rate (`/auth/device`, `/auth/token`, `/auth/refresh`, `/auth/revoke`)
- Token revocation hits (denylist matches)
- Active session count per user (distribution)
- Session revoke propagation latency (p95)

## 18.3 SLOs and Alert Thresholds

| SLO | Target | Alert Trigger |
|-----|--------|---------------|
| API latency (p95) | < 200ms | p95 > 300ms for 10m |
| API availability | 99.5% monthly | < 99.0% rolling 1h |
| Error rate | < 1% | > 2% for 5m |
| Auth failure ratio | < 5% | > 10% for 10m |
| Session revocation propagation | p95 < 60s | p95 > 120s for 10m |
| Email delivery success | > 99% | < 97% for 15m |

## 18.4 Health Check

```go
func (h *HealthHandler) Health(c *gin.Context) {
    if err := h.db.Ping(); err != nil {
        c.JSON(503, gin.H{"status": "unhealthy"})
        return
    }
    c.JSON(200, gin.H{"status": "ok", "version": Version})
}
```

---
