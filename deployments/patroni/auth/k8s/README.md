# STUB manifests for Auth Postgres HA on k3s.

> Roadmap (all Postgres services): [`../../README.md`](../../README.md) · Architecture: [`confluence/architecture/PATRONI.md`](../../../../confluence/architecture/PATRONI.md)

Apply order:

```bash
kubectl apply -f namespace.yaml
kubectl apply -f etcd.yaml
# Create secret before Patroni (use strong values — do not commit):
# kubectl -n eh-patroni-auth create secret generic patroni-auth-secrets \
#   --from-literal=superuser-password='CHANGE_ME' \
#   --from-literal=replication-password='CHANGE_ME' \
#   --from-literal=app-password='CHANGE_ME'
kubectl apply -f patroni-statefulset.yaml
kubectl apply -f haproxy-service.yaml
```

Point Auth chart/deployment:

```yaml
env:
  - name: DB_HOST
    value: auth-pg-haproxy.eh-patroni-auth.svc.cluster.local
  - name: DB_PORT
    value: "5432"
  - name: DB_NAME
    value: eventhorizon
  - name: DB_USER
    valueFrom:
      secretKeyRef:
        name: event-horizon-auth-secrets
        key: POSTGRES_USER
  - name: DB_PASSWORD
    valueFrom:
      secretKeyRef:
        name: event-horizon-auth-secrets
        key: POSTGRES_PASSWORD
```

**Not production-ready until:** PVC + storageClass, real Patroni image/config, etcd HA, network policies, and a failover drill.
