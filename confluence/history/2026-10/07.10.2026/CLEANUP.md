All files must be in special places. Frontend in frontend, backend in backend. sh files in script file folder, we must perform a cleanup, minding the dependency ways in files. If you switch places in them, placing in special forlder - be careful, you might ruin some ci\cd or else.

## Decision log (08.10.2026)

### Scripts — moved off root
| Was (root) | Now |
|------------|-----|
| `start.sh`, `restart.sh`, `persistence.sh`, `check-prod.sh` | `scripts/ops/` |
| `metrics_*.sh`, `load_and_profile.sh` | `scripts/metrics/` |
| `loadtest.js`, `loadtest_balancer.js` | `scripts/loadtest/` (legacy; prefer `deployments/k6/`) |

Scripts `cd` to repo root via `REPO_ROOT`. Prefer `make deploy` over legacy `scripts/ops/start.sh`.

### Dockerfiles — **stay at root**
`Makefile` / CI use `-f Dockerfile.<svc>.bin .` with build context `.` and `COPY services/...`. Moving them would break every docker build. Keep.

### Frontend isolation
- Games: `frontend/src/components/Games/<Game>/` now holds UI **and** zustand store (Flappy, Towers/Hexagon/Memonia).
- Shop / Inventory stores co-located under `components/Shop/`, `components/Inventory/`.
- Shared chrome stays in `components/Layout`, `Common`, `ui`.
