# Litmus Edge data source for Grafana

Real-time data streaming from [Litmus Edge](https://litmus.io) into Grafana. The plugin subscribes to NATS topics and pushes live frames into dashboard panels — no polling, no historical queries.

![Litmus Data Source](https://github.com/litmusautomation/edge-datasource/raw/main/img/le-datasource.gif)

## Quick start

**Grafana v12.2+** and **Litmus Edge v3.16+** required.

### Running inside Litmus Edge

When deployed as a container on Litmus Edge, the plugin works out of the box — it auto-detects the host from the Docker bridge network and connects to NATS without credentials.

[Add the data source](https://grafana.com/docs/grafana/latest/datasources/add-a-data-source/), select **Litmus Edge**, and click **Save & test**. That's it.

> Optionally, provide an **API token** to enable topic autocomplete in the query editor.

### Connecting to an external Litmus Edge

Toggle **External Litmus Edge** on and fill in:

| Field | Description |
| --- | --- |
| **Hostname** | IP or hostname of the Litmus Edge instance (e.g. `172.17.0.1`) |
| **Access Account Token** | Token with [NATS Proxy](https://docs.litmus.io/litmusedge/product-features/system/access-control/tokens#nats-proxy) read access |

The plugin connects to NATS on port 4222. Any `:port` in the hostname field is stripped for the NATS connection but used for the DeviceHub API (e.g. `172.17.0.1:8443` when HTTPS runs on a non-default port).

Click **Save & test** — you should see "Connected to the Edge".

## Usage

Add a panel, pick the Litmus Edge data source, and enter a NATS topic:

```
devicehub.alias.demo.bearing_temperature
```

The plugin subscribes and pushes frames once per second. Each query row handles one topic — add more rows for multiple topics.

**Template variables** work in the topic field (`$site.$area.$line.$sensor`) and resolve before each query.

### Data types

| Payload | Grafana type |
| --- | --- |
| Number | Float64 |
| String | String |
| Boolean | Boolean |
| Null | Nullable |
| JSON array / nested object | JSON (use "Extract Fields" transformation) |

Every frame includes a `Time` field sourced from the payload's `timestamp` (Unix ms) or the arrival time. DeviceHub messages also expose metadata: `tagName`, `deviceName`, `deviceId`, `datatype`, `description`, `registerId`.

### Labels in legends

Use label variables in **Field > Standard options > Display name**:

```
${__field.labels.deviceName}.${__field.labels.tagName}
```

## Limitations

- **Live data only** — no historical queries. Use a time-series database for that.
- **No wildcard topics** — `*` and `>` are not supported.
- **One topic per query** — add multiple query rows for multiple topics.
- **1 s resolution** — messages are batched once per second.
- **10 000-message buffer** — excess messages are dropped with a log warning.

## Troubleshooting

| Problem | What to check |
| --- | --- |
| **Save & test fails** (inside LE) | Can the container reach the host on port 4222? Try switching to External mode. |
| **Save & test fails** (external) | Is the host reachable on port 4222? Is the NATS Proxy enabled? Is the token valid? |
| **No data** | Topic must be an exact NATS subject — no wildcards. Verify the device is publishing. Check Grafana logs for `"Topic not found"`. |
| **Stale data after reconnect** | NATS buffers messages while disconnected. Refresh the dashboard. |
| **"Messages dropped"** | Topic exceeds 10 000 msg/s. Subscribe to a more specific topic. |

### Rotating credentials

Tokens are stored in Grafana's [secure JSON data](https://grafana.com/docs/grafana/latest/administration/provisioning/#datasources) and never sent to the browser after setup. Edit the data source, enter the new token, and Save & test.

## Contributing

See [CONTRIBUTING.md](https://github.com/litmusautomation/edge-datasource/blob/main/CONTRIBUTING.md).
