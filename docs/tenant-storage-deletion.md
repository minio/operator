# When one-time storage is required

There are times when we need to do tests, and once the tests are done, we need to clean up the storage immediately. We can do the following configuration

```$xslt
 - name: pool-0
   reclaimStorage: true
```

When a tenant is deleted, the associated pvc is also deleted immediately.