Consul Alerting
================
[![Build Status](https://travis-ci.org/kyhavlov/consul-alerting.svg?branch=master)](https://travis-ci.org/kyhavlov/consul-alerting)

This project provides a daemon to run alongside Consul and alert on health check failures. It can be configured to watch only local service and node health checks, or to use the catalog to monitor all services/checks. It distributes the alerting load by acquiring individual locks on the nodes/services it is monitoring, allowing daemons on different nodes to share the work and to pick up monitoring for one another in the event of node failure.

Usage
-----

### Command Line
To run the daemon, pass the `-config` flag for the config file location. If a config file is not specified, the default configuration settings will be used and alerts will be logged on the `stdout` handler.

`consul-alerting [--help] -config=/path/to/config.hcl`

### Configuration File(s)
The Consul Alerting configuration files are written in [HashiCorp Configuration Language (HCL)][HCL]. By proxy, this means the Consul Alerting configuration file is JSON-compatible. For more information, please see the [HCL specification][HCL].

##### Example Config
```hcl
consul_address = "localhost:8500"

node_watch = "local"
service_watch = "global"
change_threshold = 30
log_level = "info"

service "redis" {
  change_threshold = 15
  distinct_tags = true
}

service "elasticsearch" {
  distinct_tags = true
  ignored_tags = ["master", "client"]
}

handlers {
  stdout "log" {
    log_level = "warn"
  }
  email "admin" {
    recipients = ["admin@example.com"]
  }
}
```

#### Global Options

|       Option       | Description |
| ------------------ |------------ |
| `consul_address`   | The address of the consul agent to connect to. Defaults to `localhost:8500`.
| `token`            | The [Consul API token][Consul ACLs]. There is no default value.
| `node_watch`       | The setting to use for discovering nodes. If set to `local`, only the local node's health will be watched. If set to `global`, all nodes in the catalog will be watched. Defaults to `local`.
| `service_watch`    | The setting to use for discovering services. If set to `local`, only services on the local node will be watch. If set to `global`, all services in the catalog will be watched. Defaults to `local`.
| `change_threshold` | The time (in seconds) that a check must be in a failing state before alerting. Defaults to 60.
| `log_level`        | The logging level to use. Defaults to `info`.

#### Service Options
The following options can be specified in a service block:

|       Option       | Description |
| ------------------ |------------ |
| `change_threshold` | The time (in seconds) that this service must be in a failing state before alerting. Defaults to the global `change_threshold`.
| `distinct_tags`    | Treat every tag registered as a distinct service, and specify the tag when sending alerts about the failing service. Defaults to false.
| `ignored_tags`     | Tags to ignore when using `distinct_tags`. Useful when excluding generic tags like "master" that are spread across multiple clusters of the same service.

#### Handler Options
Handlers must have `enabled = true` in order to be active.

**stdout**

|       Option       | Description |
| ------------------ |------------ |
| `log_level`        | The log level to log alerts to. Defaults to "warn".

**email**

|       Option       | Description |
| ------------------ |------------ |
| `recipients`       | A list of email addresses to send alerts to.

**pagerduty**

|       Option       | Description |
| ------------------ |------------ |
| `service_key`      | The PagerDuty api key to use for alerting.
| `max_retries`      | The maximum number of times to retry after a request failure when alerting. Defaults to 0.

#### Example log output:
```
[Aug 21 19:15:19]  INFO Handler 'stdout' enabled with loglevel warn
[Aug 21 19:15:20]  INFO Running in local mode, monitoring node consul's services
[Aug 21 19:15:20]  INFO Waiting to acquire lock on node consul...
[Aug 21 19:15:20]  INFO Service found: consul, tags: []
[Aug 21 19:15:20]  INFO Service found: nginx, tags: [gamma delta]
[Aug 21 19:15:20]  INFO Service found: redis, tags: [alpha beta]
[Aug 21 19:15:20]  INFO Waiting to acquire lock on service nginx...
[Aug 21 19:15:20]  INFO Waiting to acquire lock on service redis (tag: alpha)...
[Aug 21 19:15:20]  INFO Waiting to acquire lock on service redis (tag: beta)...
[Aug 21 19:15:20]  INFO Waiting to acquire lock on service consul...
[Aug 21 19:15:20]  INFO Acquired lock for node consul
[Aug 21 19:15:20]  INFO Acquired lock for service nginx
[Aug 21 19:15:20]  INFO Acquired lock for service consul
[Aug 21 19:15:20]  INFO Acquired lock for service redis (tag: alpha)
[Aug 21 19:15:20]  INFO Acquired lock for service redis (tag: beta)
[Aug 21 19:15:30]  WARN service nginx is now warning (Unhealthy nodes: [consul])
[Aug 21 19:15:46]  WARN service nginx is now critical (Unhealthy nodes: [consul])
[Aug 21 19:15:58]  WARN node consul is now warning (Failing checks: [memory usage Service 'nginx' check Service 'redis' check])
[Aug 21 19:15:59]  WARN service redis (tag: alpha) is now warning (Unhealthy nodes: [consul])
```

[HCL]: https://github.com/hashicorp/hcl "HashiCorp Configuration Language (HCL)"
[Consul ACLs]: https://www.consul.io/docs/internals/acl.html "Consul ACLs"
