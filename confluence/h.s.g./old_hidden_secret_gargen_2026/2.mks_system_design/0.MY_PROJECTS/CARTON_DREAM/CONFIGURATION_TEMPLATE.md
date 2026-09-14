# Configuration Template

Copy these values to your `.env` file or Kubernetes secrets.

## Required Environment Variables

```bash
# JWT
JWT_SECRET=your_super_secret_jwt_key_min_32_characters_long

# Telegram Bot
TELEGRAM_API_KEY=your_telegram_bot_token_from_botfather

# DeepSeek AI
DEEPSEEK_API_KEY=sk-your_deepseek_api_key

# SMTP
SMTP_USER=your_email@mail.ru
SMTP_PASSWORD=your_app_password
MANAGER_EMAIL=manager@mail.ru

# Boosty
BOOSTY_API_KEY=your_boosty_api_key
BOOSTY_WEBHOOK_SECRET=your_webhook_secret

# Admin
ADMIN_SECRET=your_admin_secret_key
```

## How to Get API Keys

1. **Telegram Bot**: https://t.me/botfather → `/newbot`
2. **DeepSeek**: https://platform.deepseek.com → API Keys
3. **Mail.ru**: Settings → Security → App Passwords
4. **Boosty**: Contact Boosty support (API in development)

