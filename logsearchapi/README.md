# Log Search API Server for MinIO

## Development setup

1. Start Postgresql server in container:

```shell
docker run --rm -it -e "POSTGRES_PASSWORD=example" -p 5432:5432 postgres:13-alpine -c "log_statement=all"
```

2. Start logsearchapi server:

```shell
export LOGSEARCH_PG_CONN_STR="postgres://postgres:example@localhost/postgres"
export LOGSEARCH_AUDIT_AUTH_TOKEN=xxx
export LOGSEARCH_QUERY_AUTH_TOKEN=yyy
export LOGSEARCH_DISK_CAPACITY_GB=5
go build && ./logsearchapi
```

3. Minio setup:

```shell
mc admin config set myminio audit_webhook:1 'endpoint=http://localhost:8080/api/ingest?token=xxx'

mc admin service restart myminio
```

4. Sample search/list queries:

```shell
curl -v "http://localhost:8080/api/query?token=yyy&q=raw&pageNo=0&pageSize=10&timeStart=2020-11-04T22:26:12.732402319Z"
```
