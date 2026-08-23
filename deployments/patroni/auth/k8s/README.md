# STUB manifests for Auth Postgres HA on k3s.

Apply order:

```bash
kubectl apply -f namespace.yaml
kubectl apply -f etcd.yaml
# Create secret before Patroni (optional in stub):
# kubectl -n eh-patroni-auth create secret generic patroni-auth-secrets \
#   --from-literal=superuser-password=patroni \
#   --from-literal=replication-password=replicator
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
```

**Not production-ready until:** PVC + storageClass, real Patroni image/config, etcd HA, network policies, and a failover drill.
