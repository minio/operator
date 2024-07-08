# Development

## Deploy operator in K8s

With a Kubernetes cluster (local you can use Kind) in `./` run:

```
kubectl apply -k resources/ 
kubectl apply -k resources/base/
```

That will deploy an operator with a defined image.

## Deploy operator in K8s with custom image using Kind:

Build an image like:

```
make build TAG="minio/operator:<YOUR_TAG>" && kind load docker-image minio/operator:<YOUR_TAG>
```

And then use your `TAG` in the `minio-operator` deployment.