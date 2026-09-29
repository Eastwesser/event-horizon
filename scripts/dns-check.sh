#!/usr/bin/env bash
# Diagnose host DNS for docker push / go module downloads.
# ICMP (ping) can work while UDP :53 fails — they are different paths.
#
# Known EH failure mode (Amnezia / amn0 VPN):
#   docker push → lookup registry-1.docker.io on 8.8.8.8:53:
#   write udp 10.8.x.x:…→8.8.8.8:53: write: operation not permitted
# Cause: /etc/resolv.conf points only at 8.8.8.8; VPN blocks UDP to Google DNS.
# dockerd (Go) reads resolv.conf directly and fails; glibc apps via 127.0.0.53 often still work.
set -euo pipefail

echo "== resolv.conf =="
ls -la /etc/resolv.conf 2>/dev/null || true
cat /etc/resolv.conf 2>/dev/null || true
echo

echo "== UDP DNS probes (EPERM = VPN/firewall block) =="
python3 - <<'PY'
import socket
for host in ("127.0.0.53", "1.1.1.1", "8.8.8.8", "9.9.9.9"):
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    s.settimeout(2)
    try:
        # minimal A query for example.com
        q = (
            b"\x12\x34\x01\x00\x00\x01\x00\x00\x00\x00\x00\x00"
            b"\x07example\x03com\x00\x00\x01\x00\x01"
        )
        s.sendto(q, (host, 53))
        data, _ = s.recvfrom(512)
        print(f"  UDP {host}:53  OK  ({len(data)} bytes)")
    except PermissionError as e:
        print(f"  UDP {host}:53  EPERM  ← docker push will fail if resolv.conf uses this NS")
        print(f"             {e}")
    except Exception as e:
        print(f"  UDP {host}:53  FAIL  {type(e).__name__}: {e}")
    finally:
        s.close()
PY
echo

echo "== getent hosts =="
for host in proxy.golang.org registry-1.docker.io github.com; do
  echo -n "  $host: "
  getent hosts "$host" 2>&1 | head -1 || echo "(fail)"
done
echo

echo "== curl registry (TLS) =="
curl -4 -sS -o /dev/null -w "  registry-1.docker.io HTTP %{http_code}\n" \
  --connect-timeout 5 https://registry-1.docker.io/v2/ || echo "  curl failed"
echo

echo "== VPN / iface hint =="
ip -4 addr show amn0 2>/dev/null | sed 's/^/  /' || echo "  (no amn0 — Amnezia VPN not up)"
echo

echo "Hints:"
echo "  • Local redeploy does NOT need docker push — compose uses the local"
echo "    eastwesser/<svc>:latest tag after: bash scripts/rebuild-services.sh <svc>"
echo "    then: docker compose --env-file .env -f deployments/docker-compose.cluster.yml up -d <svc>…"
echo "  • Fix docker push DNS (pick one):"
echo "      sudo ln -sf /run/systemd/resolve/stub-resolv.conf /etc/resolv.conf"
echo "      # or: echo 'nameserver 1.1.1.1' | sudo tee /etc/resolv.conf"
echo "      # or: disconnect Amnezia VPN briefly, push, reconnect"
echo "  • Verify: bash scripts/dns-check.sh   then   docker push eastwesser/gateway:latest"
echo "  • Go modules: export GOPROXY=https://proxy.golang.org,direct"
echo "  • Rebuild scripts set GOMODCACHE=\$HOME/go/pkg/mod"
