# 18. Monitoring and Observability

## 18.1 Monitoring Stack

| Component | Tool |
|-----------|------|
| Logging | zerolog (JSON) |
| Log Aggregation | Fly.io logs |
| Metrics | Fly.io dashboard |
| Health Check | `/health` endpoint |

## 18.2 Key Metrics

- Request count by endpoint
- Response time percentiles
- Error rate
- Bookings created/cancelled
- Email delivery success rate

## 18.3 Health Check

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
