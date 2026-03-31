# Litmus Edge data source for Grafana

Stream live operational data from [Litmus Edge](https://litmus.io) directly into Grafana dashboards. The plugin subscribes to NATS topics on the edge and pushes frames into panels in real time — no polling, no historical database required.

![Litmus Edge query editor and live panel](https://github.com/litmusautomation/edge-datasource/blob/main/src/img/query-editor-screenshot.png?raw=true)

## Installation

Choose one of these deployment paths:

### Option 1: Use the prebuilt `litmus-grafana` image

Use the bundled Grafana image when you want the fastest path to a working setup:

```bash
docker run -p 3000:3000 \
  litmusedge.azurecr.io/litmus-grafana:latest
```

The image already includes the Litmus Edge data source plugin and provisions it automatically. See [litmus-grafana/README.md](https://github.com/litmusautomation/edge-datasource/blob/main/litmus-grafana/README.md) for environment variables and external-connection setup.

### Option 2: Install the plugin in your existing Grafana instance

Download a release zip from GitHub Releases and install the plugin with Grafana CLI or by extracting it into your Grafana plugins directory.

Example with Grafana CLI:

```bash
grafana cli \
  --pluginUrl https://github.com/litmusautomation/edge-datasource/releases/latest/download/litmus-edge-datasource.zip \
  plugins install litmus-edge-datasource
```

After installation, restart Grafana and add or provision a data source of type `litmus-edge-datasource`.

Provisioning example for inside-LE mode:

```yaml
apiVersion: 1

datasources:
  - name: Litmus Edge
    type: litmus-edge-datasource
    access: proxy
    jsonData:
      externalEdge: false
      gatewayIp: ${EDGE_DOCKER_GATEWAY_IP}
    secureJsonData:
      apiToken: ${EDGE_API_TOKEN}
```

Provisioning example for external mode:

```yaml
apiVersion: 1

datasources:
  - name: Litmus Edge
    type: litmus-edge-datasource
    access: proxy
    jsonData:
      externalEdge: true
      hostname: ${EDGE_HOSTNAME}
      natsProxyPort: ${EDGE_NATS_PROXY_PORT}
    secureJsonData:
      token: ${EDGE_ACCESS_ACCOUNT_TOKEN}
      apiToken: ${EDGE_API_TOKEN}
```

Note: the plugin is signed for localhost root URLs. If your Grafana instance uses a different `root_url`, either align it with one of the signed localhost URLs or explicitly allow the plugin with `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=litmus-edge-datasource`.

## Quick start

**Grafana v12.2+** and **Litmus Edge v3.16+** required.

### Running inside Litmus Edge

Leave **Remote Connection** off. [Add the data source](https://grafana.com/docs/grafana/latest/datasources/add-a-data-source/), select **Litmus Edge**, confirm **Edge Docker Gateway IP** (default `10.30.50.1`), and click **Save & test**.

The plugin does not auto-detect this address. Update it if your Litmus Edge instance uses a different Docker gateway IP.

> Optional but recommended: add an **API Token** to enable topic discovery in the query editor.

| Field                      | Description                                                                                                                                    |
| -------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| **Edge Docker Gateway IP** | IP used to reach Litmus Edge from the Grafana container. Default: `10.30.50.1`. Update it if this instance uses a different Docker gateway IP. |
| **API Token**              | Optional, but recommended for topic discovery in the query editor.                                                                             |

### Connecting to an external Litmus Edge

In the data source settings, turn on **Remote Connection** and provide:

| Field                    | Description                                                                                                                     |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------- |
| **Litmus Edge Address**  | Hostname or IP address of your Litmus Edge instance. Add `:port` only when Litmus Edge uses a non-default port.                 |
| **NATS Proxy Port**      | Port used for live data streaming. Default: `4222`.                                                                             |
| **Access Account Token** | Token with [NATS Proxy](https://docs.litmus.io/litmusedge/product-features/system/access-control/tokens#nats-proxy) read access |
| **API Token**            | Optional, but recommended for topic discovery in the query editor.                                                              |

The datasource uses the Litmus Edge address for connectivity and the NATS Proxy port for live streaming.

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

Every frame includes a `Time` field derived from the payload's `timestamp`. The plugin accepts Unix timestamps in milliseconds or seconds, and falls back to arrival time when the timestamp is absent or not epoch-based. DeviceHub messages also carry metadata labels: `tagName`, `deviceName`, `deviceId`, `datatype`, `description`, and `registerId`.

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
| **Save & test fails** (inside LE) | Is **Edge Docker Gateway IP** correct for this Litmus Edge instance? Can the container reach Litmus Edge on port `4222`?                |
| **Save & test fails** (external)  | Is Litmus Edge reachable on the configured NATS Proxy port? Is the NATS Proxy enabled? Is the Access Account Token valid?               |
| **No data**                       | The topic must be an exact NATS subject — no wildcards. Verify the device is publishing and check Grafana logs for `"Topic not found"`. |
| **Stale data after reconnect**    | NATS may buffer messages while disconnected. Refresh the dashboard to clear stale frames.                                               |
| **"Messages dropped"**            | The topic exceeds the 10 000-message buffer. Subscribe to a more specific subject.                                                      |

### Rotating credentials

Tokens are stored in Grafana's [secure JSON data](https://grafana.com/docs/grafana/latest/administration/provisioning/#datasources) and are never exposed to the browser after initial setup. To rotate, edit the data source, enter the new token, and click **Save & test**.

## Contributing

See [CONTRIBUTING.md](https://github.com/litmusautomation/edge-datasource/blob/main/CONTRIBUTING.md).

## License

Copyright 2026 Litmus Automation, Inc.

Licensed under the [Apache License, Version 2.0](https://github.com/litmusautomation/edge-datasource/blob/main/LICENSE).
