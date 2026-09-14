# 🎁 Carton Dream - AI-Powered Gift Packaging Platform

## 📌 Project Overview

**Carton Dream** is a comprehensive platform for creating custom gift packaging (boxes, wrapping paper, cards, envelopes, and bags) powered by **AI** (DeepSeek) and built with **Go microservices** + **Telegram Bot** + **Angular Frontend**.

### 🎯 Key Features
- 🤖 **AI-Powered Design**: DeepSeek generates creative briefs and technical specifications
- 💬 **Telegram Bot**: Natural language order creation with voice message support
- 💳 **Subscription-Based**: Flexible plans (10/100/1000 objects per month)
- 📊 **Order Tracking**: Real-time status updates from pending to completed
- 🔔 **Auto-Notifications**: Manager alerts via email when orders are created
- 🏗️ **Microservices Architecture**: Scalable, containerized Go services
- 🐰 **Event-Driven**: RabbitMQ for asynchronous communication

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    NGINX (Port 80/443)                       │
│                      API Gateway + SSL                       │
└────────┬────────────────────────────┬─────────────────────────┘
         │                            │
    ┌────▼─────┐                 ┌────▼────────┐
    │ Angular  │                 │ Telegram    │
    │ Frontend │                 │ Bot (Go)    │
    │ :4200    │                 │ + DeepSeek  │
    └──────────┘                 └─────────────┘
         │                            │
         └──────────┬─────────────────┘
                    │
    ┌───────────────▼──────────────────────────────┐
    │         Microservices (gRPC)                 │
    ├──────────────────────────────────────────────┤
    │  • Auth        :50051  (JWT, Users)          │
    │  • Profile     :50052  (Subscriptions, LLM)  │
    │  • Orders      :50053  (Order Management)    │
    │  • Payment     :50054  (Boosty Integration)  │
    │  • Notification:50055  (SMTP, RabbitMQ)      │
    └──────────────────────────────────────────────┘
                    │
    ┌───────────────▼──────────────────────────────┐
    │         Infrastructure Layer                 │
    ├──────────────────────────────────────────────┤
    │  • PostgreSQL (5 databases)                  │
    │  • RabbitMQ (Message Broker)                 │
    │  • DeepSeek API (AI Generation)              │
    │  • Boosty API (Payments)                     │
    │  • SMTP Mail.ru (Notifications)              │
    └──────────────────────────────────────────────┘
```

---

## 🔄 Complete User Journey

### 1. Registration & Authentication
User visits website → Registers → Receives Telegram bot link

### 2. Order Creation (Telegram Bot)
```
User: "Хочу красивую коробку с цветами для дня рождения"
  ↓
Bot → DeepSeek AI: Generate detailed creative brief (10-20s)
  ↓
Bot shows brief → User approves or requests changes
  ↓
Bot → DeepSeek AI: Generate technical specification (15-25s)
  ↓
Bot shows TZ → User approves or requests changes
  ↓
Order created in database (status: "pending")
```

### 3. Manager Notification
```
Orders Service → RabbitMQ: Publish "order.created" event
  ↓
Notification Service ← RabbitMQ: Consume event
  ↓
Notification Service → SMTP: Send email to manager@mail.ru
  ↓
Manager receives: Full order details, brief, and TZ
```

### 4. Subscription & Payment
```
User clicks "Подписка" → Selects plan (10/100/1000)
  ↓
Payment Service → Boosty API: Create payment link
  ↓
User pays on Boosty
  ↓
Boosty Webhook → Payment Service: Payment confirmed
  ↓
Payment Service → RabbitMQ: Publish "payment.completed"
  ↓
Profile Service: Update subscription quota
  ↓
Orders Service: Update order status to "processing"
```

### 5. Production & Delivery
```
Manager → Factory: Sends technical specification
  ↓
Factory produces packaging
  ↓
Manager updates order status: "processing" → "ready" → "completed"
  ↓
User tracks status in personal account (Angular)
  ↓
Quota decrements when order status = "completed"
```

---

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose 2.20+
- Go 1.21+ (for local development)
- Node.js 18+ (for frontend)
- API Keys: Telegram Bot, DeepSeek, Boosty

### 1. Clone Repository
```bash
cd /home/denismatveev/Downloads/GOLANG_DREAMS/only_pets/carton-dream
```

### 2. Configure Environment
```bash
cp .env.example .env
nano .env  # Fill in your API keys
```

**Required Variables:**
```bash
# Essential
JWT_SECRET=your_jwt_secret_min_32_chars
TELEGRAM_API_KEY=your_telegram_bot_token
DEEPSEEK_API_KEY=sk-your_deepseek_key

# SMTP (for manager notifications)
SMTP_USER=your_email@mail.ru
SMTP_PASSWORD=your_app_password
MANAGER_EMAIL=manager@mail.ru

# Optional (will be needed for production)
BOOSTY_API_KEY=your_boosty_key
BOOSTY_WEBHOOK_SECRET=your_webhook_secret
```

### 3. Start All Services
```bash
# Start entire stack
docker-compose up -d

# View logs
docker-compose logs -f

# Check service health
docker-compose ps
```

### 4. Verify Services
```bash
# Health checks
curl http://localhost/health              # Nginx
curl http://localhost:8081/health         # Auth service
curl http://localhost:15672               # RabbitMQ UI (carton_user/carton_pass)

# Database check
docker exec carton-db-auth psql -U carton_user -d auth_db -c "SELECT COUNT(*) FROM users;"
```

### 5. Test Telegram Bot
```bash
# Find your bot on Telegram
# Send: /start
# Bot should respond with welcome message

# Test order creation
# Send: /order
# Enter: "Хочу красивую коробку для подарка"
# Wait for AI-generated brief
```

---

## 📂 Project Structure

```
carton-dream/
├── backend/
│   ├── auth/              # Authentication service (JWT, users)
│   ├── profile/           # Profile & subscription management
│   ├── orders/            # Order management service
│   ├── payment/           # Payment & Boosty integration
│   ├── notification/      # Email notifications (SMTP + RabbitMQ)
│   ├── pubsub/            # RabbitMQ & Kafka message brokers
│   └── tools/             # Shared gRPC definitions & utilities
├── frontend/              # Angular web application
├── pompon-bot-golang/     # Telegram bot with DeepSeek integration
├── nginx/                 # Nginx configuration
├── docker-compose.yml     # Full stack orchestration
├── .env.example           # Environment variables template
├── ARCHITECTURE.md        # Detailed architecture documentation
├── DEPLOYMENT.md          # Deployment & operations guide
└── README.md             # This file
```

---

## 🗄️ Database Structure

### 5 PostgreSQL Databases

| Database | Port | Purpose |
|----------|------|---------|
| auth_db | 5432 | Users, JWT tokens, phone verification |
| orders_db | 5433 | Orders, briefs, technical specifications |
| payments_db | 5434 | Payments, Boosty transactions |
| profile_db | 5435 | Subscriptions, quotas, LLM models |
| pompon_db | 5436 | Telegram bot data (categories, products) |

### Key Tables

**orders** (Enhanced)
```sql
id, user_id, telegram_user_id, 
user_wish, brief, technical_specification,
status (pending/processing/ready/completed/cancelled),
created_at, updated_at
```

**subscriptions**
```sql
id, user_id, plan (plan_10/plan_100/plan_1000),
total_quota, remaining_quota,
started_at, expires_at, is_active
```

**payments**
```sql
id, user_id, subscription_id, amount, currency,
provider (boosty), provider_payment_id,
status (pending/completed/failed/refunded)
```

---

## 🔗 Service Communication

### gRPC Services (Internal)
- **Auth → Profile**: Verify user tokens
- **Orders → Profile**: Check subscription quota
- **Payment → Profile**: Update subscription after payment
- **Bot → All Services**: Create orders, check status

### RabbitMQ Events (Asynchronous)
- `order.created` → Notification Service → Email to manager
- `payment.completed` → Orders Service → Update status
- `order.status_changed` → Notification Service → Notify user

### External APIs
- **DeepSeek API**: AI brief and TZ generation
- **Boosty API**: Subscription payments
- **Telegram API**: Bot communication
- **SMTP Mail.ru**: Manager email notifications

---

## 🛠️ Development

### Run Individual Service
```bash
# Example: Run auth service locally
cd backend/auth
go run cmd/authsvc/main.go

# Example: Run telegram bot locally
cd pompon-bot-golang
go run cmd/pompon/main_updated.go
```

### Database Migrations
```bash
# Orders service
docker-compose run --rm migrator-orders up

# Payment service
docker-compose run --rm migrator-payments up

# Profile service
docker-compose run --rm migrator-profile up

# Bot database
docker exec -it carton-db-bot psql -U carton_user -d pompon_db
\i /docker-entrypoint-initdb.d/001_add_brief_and_tz_to_orders.sql
```

### Rebuild After Code Changes
```bash
# Rebuild specific service
docker-compose up -d --build telegram-bot

# Rebuild all services
docker-compose up -d --build
```

---

## 🔍 Testing

### Test DeepSeek Integration
```bash
# Set API key
export DEEPSEEK_API_KEY=sk-your_key

# Test brief generation
curl -X POST https://api.deepseek.com/v1/chat/completions \
  -H "Authorization: Bearer $DEEPSEEK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek-chat",
    "messages": [
      {"role": "system", "content": "You are a gift packaging designer"},
      {"role": "user", "content": "Create a brief for a birthday gift box"}
    ]
  }'
```

### Test RabbitMQ Events
```bash
# Publish test event
docker exec carton-rabbitmq rabbitmqadmin publish \
  exchange=amq.default routing_key=order.created \
  payload='{"order_id":"123","user_wish":"test"}'

# Check notification service logs
docker-compose logs -f notification
```

### Test Email Notifications
```bash
# Trigger order creation in bot
# Check manager email inbox
# Should receive email with order details
```

---

## 📊 Monitoring

### Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f telegram-bot
docker-compose logs -f notification
```

### RabbitMQ Management
- URL: http://localhost:15672
- User: `carton_user`
- Password: `carton_pass`
- Check queues: `order.created`, `payment.completed`

### Database Queries
```bash
# View recent orders
docker exec -it carton-db-orders psql -U carton_user -d orders_db \
  -c "SELECT id, status, created_at FROM orders ORDER BY created_at DESC LIMIT 10;"

# View active subscriptions
docker exec -it carton-db-profile psql -U carton_user -d profile_db \
  -c "SELECT user_id, plan, remaining_quota FROM subscriptions WHERE is_active = true;"
```

---

## 🐛 Troubleshooting

### Bot Not Responding
```bash
# Check bot container
docker-compose logs telegram-bot | tail -20

# Verify token
curl https://api.telegram.org/bot${TELEGRAM_API_KEY}/getMe

# Restart bot
docker-compose restart telegram-bot
```

### DeepSeek API Errors
```bash
# Check API key is set
docker exec carton-telegram-bot env | grep DEEPSEEK

# Test API connectivity
docker exec carton-telegram-bot curl -I https://api.deepseek.com
```

### Email Not Sending
```bash
# Check notification service logs
docker-compose logs notification | grep -i error

# Verify SMTP settings
docker exec carton-notification env | grep SMTP

# Test SMTP connection
telnet smtp.mail.ru 587
```

### Database Connection Issues
```bash
# Check database health
docker-compose ps | grep db

# Test connection
docker exec carton-db-auth pg_isready -U carton_user

# View connection logs
docker-compose logs db_auth
```

---

## 📚 Documentation

- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Detailed system architecture
- **[DEPLOYMENT.md](DEPLOYMENT.md)** - Production deployment guide
- **[pompon-bot-golang/TELEGRAM_BOT_GUIDE.md](pompon-bot-golang/TELEGRAM_BOT_GUIDE.md)** - Telegram bot documentation

---

## 🔒 Security Best Practices

### ✅ For Production

1. **Change Default Passwords**
   ```bash
   # Update in .env
   POSTGRES_PASSWORD=<strong-random-password>
   JWT_SECRET=<min-32-char-random-string>
   ```

2. **Enable SSL/TLS**
   ```bash
   # Generate Let's Encrypt certificates
   sudo certbot certonly --standalone -d your-domain.com
   
   # Copy to nginx/ssl/
   cp /etc/letsencrypt/live/your-domain.com/*.pem nginx/ssl/
   ```

3. **Firewall Rules**
   ```bash
   # Close database ports to external access
   sudo ufw deny 5432:5436/tcp
   
   # Allow only necessary ports
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   ```

4. **Secrets Management**
   - Don't commit `.env` to Git
   - Use environment-specific configs
   - Rotate API keys regularly

---

## 🚀 Roadmap

### Phase 1: MVP (Completed ✅)
- [x] Telegram bot with DeepSeek AI
- [x] Order workflow (brief → TZ → notification)
- [x] Manager email notifications
- [x] Docker Compose orchestration
- [x] RabbitMQ event-driven architecture

### Phase 2: Production Ready (In Progress 🚧)
- [ ] Boosty payment integration
- [ ] Subscription quota management
- [ ] Angular frontend (auth, account, subscription)
- [ ] gRPC inter-service communication
- [ ] SSL/TLS for production

### Phase 3: Enhancements (Planned 📋)
- [ ] Voice message support (Whisper API)
- [ ] Image generation (DALL-E/Midjourney)
- [ ] Admin panel for managers
- [ ] Analytics dashboard
- [ ] Multi-language support

### Phase 4: Scale (Future 🔮)
- [ ] Kubernetes deployment (k3s)
- [ ] Kafka for high-throughput events
- [ ] Redis for caching
- [ ] Horizontal pod autoscaling
- [ ] Service mesh (Istio/Linkerd)

---

## 📞 Support

- **Email**: support@carton-dream.com
- **Telegram**: @carton_dream_support
- **Issues**: [GitHub Issues](#)

---

## 👥 Contributors

- **Denis Matveev** - Project Lead & Full Stack Development

---

## 📄 License

Proprietary - All Rights Reserved

---

## 🎉 Acknowledgments

- **DeepSeek AI** - For powerful LLM capabilities
- **Telegram** - For bot platform
- **Go Community** - For excellent tooling
- **Docker** - For containerization

---

**Built with ❤️ in Russia** 🇷🇺

**Status**: 🚧 MVP Ready - Production Deployment In Progress  
**Last Updated**: 2025-11-15  
**Version**: 1.0.0

