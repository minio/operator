### 1) Install the MinIO Operator via Krew Plugin

Run the following commands to install the MinIO Operator and Plugin using the Kubernetes ``krew`` plugin manager:

```sh
kubectl krew update
kubectl krew install minio
```

See the ``krew`` [installation documentation](https://krew.sigs.k8s.io/docs/user-guide/setup/install/) for instructions
on installing ``krew``.

Run the following command to verify installation of the plugin:

```sh
kubectl minio version
```

As an alternative to `krew`, you can download the `kubectl-minio` plugin from
the [Operator Releases Page](https://github.com/minio/operator/releases). Download the `kubectl-minio` package
appropriate for your operating system and extract the contents as `kubectl-minio`. Set the `kubectl-minio` binary to be
executable (e.g. `chmod +x`) and place it in your system `PATH`.

For example, the following code downloads the latest stable version of the MinIO Kubernetes Plugin and installs it to
the system ``$PATH``. The example assumes a Linux operating system:

```sh
wget -qO- https://github.com/minio/operator/releases/latest/download/kubectl-minio_linux_amd64_v1.zip | sudo bsdtar -xvf- -C /usr/local/bin
sudo chmod +x /usr/local/bin/kubectl-minio
```

Run the following command to verify installation of the plugin:

```sh
kubectl minio version
```

Run the following command to initialize the Operator:

```sh
kubectl minio init
```

Run the following command to verify the status of the Operator:

```sh
kubectl get pods -n minio-operator
```

The output resembles the following:

```sh
NAME                              READY   STATUS    RESTARTS   AGE
console-6b6cf8946c-9cj25          1/1     Running   0          99s
minio-operator-69fd675557-lsrqg   1/1     Running   0          99s
```

The `console-*` pod runs the MinIO Operator Console, a graphical user
interface for creating and managing MinIO Tenants.

The `minio-operator-*` pod runs the MinIO Operator itself.

### 2) Access the Operator Console

Run the following command to create a local proxy to the MinIO Operator
Console:

```sh
kubectl minio proxy -n minio-operator
```

The output resembles the following:

```sh
kubectl minio proxy
Starting port forward of the Console UI.

To connect open a browser and go to http://localhost:9090

Current JWT to login: TOKENSTRING
```

Open your browser to the provided address and use the JWT token to log in
to the Operator Console.

![Operator Console](docs/images/operator-console.png)

Click **+ Create Tenant** to open the Tenant Creation workflow.
