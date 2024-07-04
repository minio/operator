# Development

## Deploy operator in K8s
With a Kubernetes cluster (local you can use Kind) in `./` run:
```
kubectl apply -k resources/ 
kubectl apply -k resources/base/
```
That will deploy an operator with a defined image.

To access the UI get the operator token from:
```
kubectl describe secrets -n minio-operator console-sa-secret | grep 'token:' | awk '{print $2}' | pbcopy
```

Once the server is running you can see the UI either by port forwarding or using kubefwd for the operator pod.

For development you can also run locally yarn in the web-app/ folder just make sure the port in `package.json` is the same as the operator console service.
```
"proxy": "http://localhost:9090/"
```

And in the web-app run:
```
yarn install
yarn start
```

## Deploy operator in K8s with custom image using Kind:

Build an image like:
```
make build TAG="minio/operator:<YOUR_TAG>" && kind load docker-image minio/operator:<YOUR_TAG>
```

And update the image from resources/base/ for `console-ui.yaml` and `deployment.yaml`
```
spec:
  replicas: 1
  selector:
    matchLabels:
      app: console
  template:
    metadata:
      labels:
        app: console
        app.kubernetes.io/instance: minio-operator-console
        app.kubernetes.io/name: operator
    spec:
      containers:
        - args:
            - ui
            - --certs-dir=/tmp/certs
          image: minio/operator:cesnietor

```

Apply resources again to see this:
```
kubectl apply -k resources/base/
```