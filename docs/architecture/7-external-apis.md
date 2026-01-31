# 7. External APIs

## 7.1 GitHub OAuth API

- **Purpose:** User authentication via GitHub Device Flow
- **Documentation:** https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#device-flow
- **Base URL:** https://github.com/login/device/code
- **Authentication:** Client ID + Client Secret
- **Required Scopes:** `read:user, user:email`

## 7.2 SendGrid Email API

- **Purpose:** Production email delivery to hotels
- **Documentation:** https://docs.sendgrid.com/api-reference/mail-send/mail-send
- **Base URL:** https://api.sendgrid.com/v3/mail/send
- **Authentication:** API Key (Bearer token)
- **Rate Limits:** Free tier: 100 emails/day

---
