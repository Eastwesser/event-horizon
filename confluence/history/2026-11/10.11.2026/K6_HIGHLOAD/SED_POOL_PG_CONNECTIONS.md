# auth, profile — 100/50
sed -i 's/poolCfg.MaxConns = 25/poolCfg.MaxConns = 100/; s/poolCfg.MinConns = 10/poolCfg.MinConns = 50/' \
  services/auth/internal/app/di.go \
  services/profile/internal/app/di.go

# billing, leaderboard — 64/32
sed -i 's/poolCfg.MaxConns = 25/poolCfg.MaxConns = 64/; s/poolCfg.MinConns = 10/poolCfg.MinConns = 32/' \
  services/billing/internal/app/di.go \
  services/leaderboard/internal/app/di.go

# payment, authors, history, notification — 32/16
sed -i 's/poolCfg.MaxConns = 25/poolCfg.MaxConns = 32/; s/poolCfg.MinConns = 10/poolCfg.MinConns = 16/' \
  services/payment/internal/app/di.go \
  services/authors/internal/app/di.go \
  services/history/internal/app/di.go \
  services/notification/internal/app/di.go