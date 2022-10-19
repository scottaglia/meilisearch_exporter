# Meilisearch Exporter

Prometheus exporter for Meilisearch.

### Configuration

Below is the command line options summary:

```bash
meiliearch_exporter --help
```

| Name                 | Description                                                                                | Default               |
|----------------------|--------------------------------------------------------------------------------------------|-----------------------|
| --web.systemd-socket | Use systemd socket activation listeners instead of port listeners (Linux only).            |                       |
| --web.listen-address | Addresses on which to expose metrics and web interface. Repeatable for multiple addresses. | :9974                 |
| --web.config.file    | [EXPERIMENTAL] Path to configuration file that can enable TLS or authentication. See [Web Configuration](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md)   |                       |
| --web.telemetry-path | Path under which to expose metrics.                                                        | /metrics              |
| --ms.uri             | HTTP API address of the Meilisearch host.                                                  | http://localhost:7700 |
| --ms.apikey          | Meilisearch API Key.                                                                       |                       |
| --ms.timeout         | Timeout for trying to get stats from Meilisearch.                                          | 5s                    |
| --log.level          | Only log messages with the given severity or above. One of: [debug, info, warn, error]     | info                  |
| --log.format         | Output format of log messages. One of: [logfmt, json]                                      | logfmt                |
| --version            | Show application version.                                                                  |                       |


The following environment variables can be used to override the above configurations

| Variable                    | Overrides    |
|-----------------------------|--------------|
| MEILISEARCH_EXPORTER_URI    | --ms.uri     |
| MEILISEARCH_EXPORTER_APIKEY | --ms.apikey  |


### Metrics

| Name                            | Description                                                                                                                                                                 | Labels |
|---------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------|
| meilisearch_up                  | Meilisearch instance is up and running                                                                                                                                      |        |
| meilisearch_total_scrapes       | Total of Meilisearch scrapes.                                                                                                                                               |        |
| meilisearch_last_update         | When the last update was made to the database                                                                                                                               |        |
| meilisearch_database_size       | Size of the database in bytes                                                                                                                                               |        |
| meilisearch_is_indexing         | If 1, the index is still processing documents and attempts to search will result in undefined behavior. If 0, the index has finished processing and you can start searching | index  |
| meilisearch_number_of_documents | Total number of documents                                                                                                                                                   | index  |
