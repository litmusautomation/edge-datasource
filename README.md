# Litmus edge data source for Grafana

The Litmus Edge data source plugin enables the visualization of real time data streaming from the edge in Grafana.

[Watch Demo](https://github.com/user-attachments/assets/934e18e8-e89d-42eb-8455-75010db3d641)

![Litmus Data Source](https://github.com/litmusautomation/edge-datasource/raw/main/img/le-datasource.gif)

## Requirements

- Grafana v12.2+
- Litmus Edge v3.16.x
- [Litmus Edge API Account](https://docs.litmus.io/litmusedge/product-features/system/access-control/tokens/create-api-account)

> [!NOTE]
> Make sure [NATS proxy](https://docs.litmus.io/litmusedge/product-features/system/access-control/tokens/create-api-account) is enabled and has read access to the topics.

## Configure the data source

[Add a new data source](https://grafana.com/docs/grafana/latest/datasources/add-a-data-source/) and select Litmus Edge. To configure the data source, you need to provide the following fields:

- **Hostname**: The hostname of the Litmus Edge instance.
- **Token**: The [token](https://docs.litmus.io/litmusedge/product-features/system/tokens/create-api-account) to authenticate with the Litmus Edge instance.

![Data Source Configuration](https://github.com/litmusautomation/edge-datasource/raw/main/img/le-datasource-config.png)

## Stream data from the edge

> **Note:** This plugin streams live data only — it does not query historical messages.

To stream data, add a new query panel and enter a **Topic** — the dot-separated
NATS subject published by Litmus Edge (e.g. `enterprise.site.area.line.cell.tag`).

![Query Configuration](https://github.com/litmusautomation/edge-datasource/raw/main/img/le-datasource-query.png)

### Supported data types

The plugin handles the following payload types automatically:

| Payload type | Grafana field type |
|-------------|-------------------|
| Number | Float64 |
| String | String |
| Boolean | Boolean |
| JSON object | Nested fields (use **Extract Fields** transformation) |

A `timestamp` field is added automatically when the payload does not include one.
For DeviceHub tags the plugin also populates labels with the tag context
(device name, data type, description, etc.).

### Template variables

Dashboard variables are supported in the topic field — for example
`$site.$area.$line.$sensor` — and are resolved before each query executes.

### Limitations

- Wildcard topics (`*`, `>`) are not allowed.
- Each query subscribes to exactly one topic.

## Development

Prerequisites: **Node.js 22+**, **Go** (see `go.mod`), **Docker** with Compose v2, and [**Mage**](https://magefile.org/) (`go install github.com/magefile/mage@latest`). The stack follows [Grafana Plugin Tools](https://grafana.com/developers/plugin-tools).

1. **Install** — `npm install`
2. **Environment** — `cp .env.example .env` and set `LITMUS_EDGE_HOSTNAME` and `LITMUS_EDGE_TOKEN` so the provisioned datasource can reach Litmus Edge when using Docker.
3. **Build** — `npm run build` (frontend into `dist/`) and `npm run Build:backend` or `mage -v build:linux` (backend binary into `dist/`).
4. **Run Grafana** — `npm run server` (or `npm run up` for detached). The dev image mounts `dist/` and `provisioning/`; use OSS with `GRAFANA_IMAGE=grafana-oss` if you prefer.
5. **Watch frontend** — in another terminal, `npm run dev` while Grafana is running for live reload of the plugin UI.

Checks before a PR: `npm run typecheck`, `npm run lint`, `npm run test:ci`. End-to-end (local, with Litmus Edge reachable via `.env`): start Grafana, then `npm run e2e:install` once and `npm run e2e`.
