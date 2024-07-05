# Tenant Storage Deletion

Deleting a tenant's storage in Kubernetes is a sensitive and controversial issue. On one hand, some users may want to
delete storage in CI/CD scenarios. On the other hand, if configured in production environments and forgotten, it risks
potential data loss.

Kubernetes does not automatically delete Persistent Volume Claims (PVCs) when a StatefulSet is deleted, due to the risk
involved. This decision is left to the user. The MinIO Operator adheres to this practice as well.

To delete a tenant's storage, you should manually delete the PVCs associated with the tenant. This can be done via
kubectl at the same time you are deleting the tenant. For example, to delete a tenant named tenant in the namespace
ns-1, run the following commands:

```bash
kubectl -n ns-1 delete tenant tenant
kubectl -n ns-1 delete pvc -l v1.min.io/tenant=tenant
```