# Modular compose overlays (Week 3 / Kozirev deploy/compose pattern)
#
# Full production-like cluster:
#   docker compose -f deployments/docker-compose.cluster.yml up -d
#
# Lightweight local deps only:
#   docker compose -f deployments/compose/core/docker-compose.yml up -d
#   docker compose -f deployments/compose/core/docker-compose.yml \
#     -f deployments/compose/auth/docker-compose.yml up -d
#   docker compose -f deployments/compose/core/docker-compose.yml \
#     -f deployments/compose/inventory/docker-compose.yml up -d
#
# App images still come from eastwesser/*:latest (bin Dockerfiles).
