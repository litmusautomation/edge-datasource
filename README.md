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

To stream data from the edge, you need to create a new query and provide the following fields:

- **Topic**: The topic name to fetch the data from.

![Query Configuration](https://github.com/litmusautomation/edge-datasource/raw/main/img/le-datasource-query.png)

> - The plugin supports topics publishing numbers, strings, boolean, and JSON objects. Use the `Extract Fields` transformation to extract the fields from the JSON object.
> - The plugin automatically adds the `timestamp` field to the query result if it is not present in the topic data.
> - The plugin automatically adds the topic context for Devicehub tags.
> - Wildcard topics are not allowed.

## Development

Prerequisites: **Node.js 20+**, **Go** (see `go.mod`), **Docker** with Compose v2, and [**Mage**](https://magefile.org/) (`go install github.com/magefile/mage@latest`). The stack follows [Grafana Plugin Tools](https://grafana.com/developers/plugin-tools).

1. **Install** — `npm install`
2. **Environment** — `cp .env.example .env` and set `LITMUS_EDGE_HOSTNAME` and `LITMUS_EDGE_TOKEN` so the provisioned datasource can reach Litmus Edge when using Docker.
3. **Build** — `npm run build` (frontend into `dist/`) and `npm run Build:backend` or `mage -v build:linux` (backend binary into `dist/`).
4. **Run Grafana** — `npm run server` (or `npm run up` for detached). The dev image mounts `dist/` and `provisioning/`; use OSS with `GRAFANA_IMAGE=grafana-oss` if you prefer.
5. **Watch frontend** — in another terminal, `npm run dev` while Grafana is running for live reload of the plugin UI.

Checks before a PR: `npm run typecheck`, `npm run lint`, `npm run test:ci`. End-to-end: start Grafana, then `npm run e2e:install` once and `npm run e2e`.
