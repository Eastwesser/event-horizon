# CI lint / delivery issues — 29.09.2026

## Fixed on `main` (this change)

### actionlint — Node 20 deprecated (6 errors)
Bumped workflow actions to Node 24 runtimes:

| Was | Now |
|-----|-----|
| `actions/checkout@v4` | `@v5` |
| `actions/setup-go@v5` | `@v6` |
| `docker/setup-buildx-action@v3` | `@v4` |
| `docker/login-action@v3` | `@v4` |

Files: `.github/workflows/main.yml` (+ synced `delivery/ci-cd/.github/workflows/*`).

### build exit code 2
Root cause was **not** Node — CI ran `make docker-build-all` without compiling the
gitignored `*-service` binaries that `Dockerfile.*.bin` COPY. Added a
`Build linux/amd64 service binaries` step (nats-hub → `nats-hub`, others → `*-service`).

### Notices
Pinned `runs-on: ubuntu-24.04` (avoids the ubuntu-latest → 26 migration notice).

## Original annotations (for history)

```
runtime "node20" deprecated — update checkout@v4, setup-go@v5,
  docker/setup-buildx-action@v3, docker/login-action@v3
build: Process completed with exit code 2
  (COPY …/analytics-service not found, etc.)
ubuntu-latest will migrate to Ubuntu 26 (Oct 2026)
```
