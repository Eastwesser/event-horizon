# Copy *.env.template → *.env (never commit secrets).
# Example:
#   cp deployments/env/auth.env.template deployments/env/auth.env
#   set -a && source deployments/env/auth.env && set +a
#   (cd services/auth && go run ./cmd/main.go)

Templates:
- auth.env.template
- billing.env.template
- inventory.env.template
