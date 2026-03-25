# Litmus Edge data source for Grafana

A Grafana datasource plugin that subscribes to NATS topics on a [Litmus Edge](https://litmus.io) instance and streams the data into dashboard panels in real time.

This is a live-only plugin. It does not query or store historical data.

[Watch Demo](https://github.com/user-attachments/assets/934e18e8-e89d-42eb-8455-75010db3d641)

![Litmus Data Source](https://github.com/litmusautomation/edge-datasource/raw/main/img/le-datasource.gif)

## Requirements

- Grafana v12.2+
- Litmus Edge v3.16+
- A [Litmus Edge Access Account](https://docs.litmus.io/litmusedge/product-features/system/access-control/tokens/create-api-account) with a token that has NATS Proxy read access
- Network connectivity from Grafana to the Litmus Edge instance on port 4222 (NATS)

> [!NOTE]
> The [NATS Proxy](https://docs.litmus.io/litmusedge/product-features/system/access-control/tokens#nats-proxy) must be enabled on the Litmus Edge instance and the Access Account token needs read access to the topics you want to stream.

## Configure the data source

[Add a new data source](https://grafana.com/docs/grafana/latest/datasources/add-a-data-source/) and select Litmus Edge.

| Field                | What to enter                              | Example                        |
| -------------------- | ------------------------------------------ | ------------------------------ |
| Hostname             | Hostname or IP of the Litmus Edge instance | `192.168.1.100`                |
| Access Account Token | Token with NATS Proxy read access          | _(stored securely by Grafana)_ |

Click "Save & test". If the connection works, you'll see "Connected to the Edge".

![Data Source Configuration](https://github.com/litmusautomation/edge-datasource/raw/main/img/le-datasource-config.png)

## Streaming data

Add a panel, pick the Litmus Edge data source, and type in a topic, the dot-separated NATS subject published by Litmus Edge:

```
devicehub.alias.demo.bearing_temperature
```

The plugin subscribes to that topic and pushes frames to the panel once per second.

![Query Configuration](https://github.com/litmusautomation/edge-datasource/raw/main/img/le-datasource-query.png)

### Data types

The plugin parses JSON payloads and maps values to Grafana field types:

| Payload type       | Grafana field type                             |
| ------------------ | ---------------------------------------------- |
| Number             | Float64                                        |
| String             | String                                         |
| Boolean            | Boolean                                        |
| Null               | Nullable field                                 |
| JSON array         | JSON                                           |
| Nested JSON object | JSON (use the "Extract Fields" transformation) |

Every frame includes a `Time` field. If the payload has a `timestamp` field (Unix milliseconds), that value is used. Otherwise the plugin uses the arrival time.

For DeviceHub (DH) tag messages, the plugin extracts metadata too: `tagName`, `deviceName`, `deviceId`, `datatype`, `description`, and `registerId`.

### Template variables

Dashboard variables work in the topic field:

```
$site.$area.$line.$sensor
```

They're resolved before each query runs, so panels update when you change a variable.

### Showing labels in legend and tooltip

DH metadata labels are preserved (for example `deviceName`, `tagName`, `deviceId`, `topic`).

In panel **Field > Standard options > Display name**, use label variables like:

```
${__field.labels.deviceName}.${__field.labels.tagName}
```

Examples:

```
${__field.labels.topic}
${__field.name} (${__field.labels.deviceId})
```

### Multiple topics

Each query row subscribes to one topic. To stream several topics, add more query rows in the same panel. They're independent, so a failure in one won't affect the others.

## Limitations

- Live data only. There is no historical query support. Use a time-series database for that.
- No wildcard topics. `*` and `>` are not supported. Each query needs an exact topic.
- One topic per query. For multiple topics, add multiple query rows.
- 1-second resolution. Messages are batched and sent to the panel once per second.
- 10,000-message buffer per topic. If a topic produces more than that between sends, the excess is dropped and a warning is logged.

## What it supports

| Feature             | Supported |
| ------------------- | --------- |
| Metrics (streaming) | Yes       |
| Template variables  | Yes       |
| Alerting            | No        |
| Annotations         | No        |
| Logs                | No        |
| Historical queries  | No        |

## Troubleshooting

### "Save & test" fails with a connection error

Check three things: can Grafana reach the Litmus Edge host on port 4222? Is the NATS Proxy enabled on the Edge instance? Is the Access Account token valid and not expired?

### Panel shows "No data"

The topic must be the exact NATS subject; wildcards won't work. Make sure the device is actually publishing on that topic. You can also check Grafana server logs for `"Topic not found"` or `"Failed to convert topic to data frame"` messages.

### Stale data after a reconnection

The plugin reconnects to NATS automatically. NATS buffers messages while disconnected, so you might see a burst of old data when it comes back. Refresh the dashboard if the panel looks off.

### "Messages dropped (buffer full)" in logs

The topic is getting more than 10,000 messages per second. Subscribe to a more specific topic, or filter at the source.

### Rotating credentials

Tokens are stored in Grafana's [secure JSON data](https://grafana.com/docs/grafana/latest/administration/provisioning/#datasources) and are never sent to the browser after initial setup. To rotate, edit the data source, enter the new token, and hit "Save & test".

## Contributing

See [CONTRIBUTING.md](https://github.com/litmusautomation/edge-datasource/blob/main/CONTRIBUTING.md) for development setup, build commands, and how to run tests.
