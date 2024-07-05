# Tenant Storage Deletion

Deleting the storage of a tenant is a controversial and delicate situation, on one hand, some users wants to delete the
storage in CI/CD sitations, but on the other hand, some users may configure this in production environments and forger
which
exposes them to potential data loss.

Kubernetes official doesn't cascade delete PVCs when a Statefulset due to the delicate nature of the operation, it's
instead
deferred to the user to decide when to delete the PVCs, MinIO Operator follows the same approach.

To delete the storage of a tenant, you can delete the PVCs associated with the tenant via kubectl same time you are
deleting a tenant, for example, when deleting a tenant called `tenant` in namespace `ns-1`, you can delete the PVCs
associated with the tenant by running the following commands:

```bash
kubectl -n ns-1 delete tenant tenant
kubectl -n ns-1 delete pvc -l v1.min.io/tenant=tenant
```