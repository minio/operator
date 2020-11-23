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

## API Documentation

### Ingest 

```
POST `/api/ingest?token=xxx`
```

This API is used to send MinIO audit logs for ingestion into the API service.

The `token` parameter is used to authenticate the request and should be equal to the `LOGSEARCH_AUDIT_AUTH_TOKEN` environment variable passed to the server.

The body must be a JSON object representing a single audit log object created by a MinIO server.

This endpoint must be configured as the audit log endpoint in the MinIO server.

### Query

```
GET /api/query?token=xxx&...
```

This API is used to query MinIO audit logs stored by the API service.

The `token` parameter is used to authenticate the request and should be equal to the `LOGSEARCH_QUERY_AUTH_TOKEN` environment variable passed to the server.

Additional query parameters specify the logs to be retrieved.

| Query parameter      | Value Description                                                                                                 | Required | Default             |
|----------------------|-------------------------------------------------------------------------------------------------------------------|----------|---------------------|
| `q`                  | `reqinfo` or `raw`.                                                                                               | Yes      | -                   |
| `timeStart`          | RFC3339 time or date. Examples: `2006-01-02T15:04:05.999999999Z07:00` or `2006-01-02`.                            | No       | Current server time |
| `timeAsc`/`timeDesc` | Flag parameter (no value); either one may be specified. Specifies result ordering.                                | No       | `timeDesc`          |
| `pageSize`           | Number of results to return per API call. Allows values between 10 and 1000.                                      | No       | `10`                |
| `pageNo`             | 0-based page number of results.                                                                                   | No       | `0`                 |
| `fp`                 | Repeatable parameter specifying key-value match filters. See the [filter parameters](#filter-parameters) section. | No       | -                   |

#### Filter Parameters

Filter parameters only work when `q=reqinfo` is set. Filter parameters allow filtering records based on key-value pattern matching. 

The format for each filter pattern is `key:value-pattern`. The `key` is the name of the field to match on can be one of `bucket`, `object`, `api_name`, `request_id`, `user_agent` or `response_status`. The value is a glob expression using `.` to signify any single character and `*` to match any number of characters. For example `bucket:photos-*` matches any bucket with a `photos-` prefix. To match a literal `.` or `*` prefix it with a `\`. To match a literal `\`, just double it: `\\`.
