# Litmus Edge data source for Grafana

Stream live operational data from [Litmus Edge](https://litmus.io) directly into Grafana dashboards. The plugin subscribes to NATS topics on the edge and pushes frames into panels in real time — no polling, no historical database required.

## Quick start

**Grafana v12.2+** and **Litmus Edge v3.16+** required.

### Running inside Litmus Edge

When deployed as a container on Litmus Edge, the plugin works out of the box — it discovers the host automatically via the Docker bridge network and connects to NATS without credentials.

[Add the data source](https://grafana.com/docs/grafana/latest/datasources/add-a-data-source/), select **Litmus Edge**, and click **Save & test**. That's it.

> Optionally, provide an **API token** to enable topic autocomplete in the query editor.

### Connecting to an external Litmus Edge

Toggle **External Litmus Edge** on and provide:

| Field                    | Description                                                                                                                     |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------- |
| **Hostname**             | IP or hostname of the Litmus Edge instance (e.g. `172.17.0.1`)                                                                  |
| **Access Account Token** | Token with [NATS Proxy](https://docs.litmus.io/litmusedge/product-features/system/access-control/tokens#nats-proxy) read access |

The plugin connects to NATS on port 4222. A `:port` suffix in the hostname is stripped for the NATS connection but preserved for DeviceHub API calls (e.g. `172.17.0.1:8443` when HTTPS runs on a non-default port).

Click **Save & test** — you should see "Connected to the Edge".

## Usage

Add a panel, select the Litmus Edge data source, and enter a NATS topic:

```
devicehub.alias.demo.bearing_temperature
```

The plugin subscribes and streams frames once per second. Each query row maps to one topic — add multiple rows to monitor several data points on the same panel.

**Template variables** are supported in the topic field (`$site.$area.$line.$sensor`) and resolve before each query, making it easy to build reusable dashboards across sites or production lines.

### Data types

| Payload                    | Grafana type                                   |
| -------------------------- | ---------------------------------------------- |
| Number                     | Float64                                        |
| String                     | String                                         |
| Boolean                    | Boolean                                        |
| Null                       | Nullable                                       |
| JSON array / nested object | JSON (use the "Extract Fields" transformation) |

Every frame includes a `Time` field derived from the payload's `timestamp` (Unix ms), falling back to arrival time when absent. DeviceHub messages also carry metadata labels: `tagName`, `deviceName`, `deviceId`, `datatype`, `description`, and `registerId`.

### Customizing legends

Use label variables in **Field > Standard options > Display name** to build meaningful series names:

```
${__field.labels.deviceName}.${__field.labels.tagName}
```

## Known limitations

- **Live data only** — no historical queries. Pair with a time-series database for historical analysis.
- **No wildcard topics** — NATS wildcards (`*`, `>`) are not supported.
- **One topic per query row** — add multiple rows for multiple topics.
- **1-second resolution** — incoming messages are batched and pushed once per second.
- **10 000-message buffer** — messages beyond the buffer cap are dropped (logged as a warning).

## Troubleshooting

| Problem                           | What to check                                                                                                                           |
| --------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| **Save & test fails** (inside LE) | Can the container reach the host on port 4222? Try switching to External mode.                                                          |
| **Save & test fails** (external)  | Is the host reachable on port 4222? Is the NATS Proxy enabled? Is the token valid?                                                      |
| **No data**                       | The topic must be an exact NATS subject — no wildcards. Verify the device is publishing and check Grafana logs for `"Topic not found"`. |
| **Stale data after reconnect**    | NATS may buffer messages while disconnected. Refresh the dashboard to clear stale frames.                                               |
| **"Messages dropped"**            | The topic exceeds the 10 000-message buffer. Subscribe to a more specific subject.                                                      |

### Rotating credentials

Tokens are stored in Grafana's [secure JSON data](https://grafana.com/docs/grafana/latest/administration/provisioning/#datasources) and are never exposed to the browser after initial setup. To rotate, edit the data source, enter the new token, and click **Save & test**.

## Contributing

See [CONTRIBUTING.md](https://github.com/litmusautomation/edge-datasource/blob/main/CONTRIBUTING.md).

## License

Copyright 2026 Litmus Automation, Inc.

Licensed under the [Apache License, Version 2.0](LICENSE).
