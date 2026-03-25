# litmus-grafana

A Grafana Docker image pre-bundled with the Litmus Edge datasource plugin and a curated set of community plugins for industrial/IoT dashboards.

## What's included

- **[Litmus Edge datasource](../README.md)** — real-time data streaming from Litmus Edge via NATS, privately signed
- **[HTML Graphics](https://grafana.com/grafana/plugins/gapit-htmlgraphics-panel/)** — custom HTML/SVG visualizations
- **[Infinity](https://grafana.com/grafana/plugins/yesoreyeram-infinity-datasource/)** — query any REST, GraphQL, or CSV endpoint
- **[Business Variable](https://grafana.com/grafana/plugins/volkovlabs-variable-panel/)** — enhanced dashboard variable UI
- **[Business Input](https://grafana.com/grafana/plugins/marcusolsson-static-datasource/)** — static/mock data for dashboards
- **[Plotly.js](https://grafana.com/grafana/plugins/nline-plotlyjs-panel/)** — Plotly.js charts

## Usage

```bash
docker run -p 3000:3000 \
  -e LITMUS_EDGE_HOSTNAME=172.17.0.1 \
  -e LITMUS_EDGE_ACCESS_ACCOUNT_TOKEN=<your-access-account-token> \
  -e LITMUS_EDGE_API_TOKEN=<your-api-token> \
  us-docker.pkg.dev/litmus-customer-facing/litmus-solutions/litmus-grafana
```

Open `http://localhost:3000` in your browser (default credentials: `admin` / `admin`).

The Litmus Edge datasource is automatically provisioned and set as the default. No manual configuration needed as long as the environment variables are set correctly.

## Environment variables

ENV GF_SERVER_ROOT_URL=http://localhost:3000
| Variable | Required | Description |
| ---------------------------------- | -------- | ------------------------------------------------ |
| `LITMUS_EDGE_HOSTNAME` | Yes | Hostname or IP of the Litmus Edge instance |
| `LITMUS_EDGE_ACCESS_ACCOUNT_TOKEN` | Yes | Access Account token with NATS Proxy read access |
| `LITMUS_EDGE_API_TOKEN` | Yes | API token for Litmus Edge REST API access |

## Plugin signature

The Litmus Edge datasource plugin is **privately signed** for localhost. The signature is valid when Grafana's `root_url` is one of:

- `http://localhost:3000` (default)
- `http://localhost:3001`
- `http://localhost:3002`
- `http://localhost:8080`
- `http://localhost:8443`

If you run Grafana behind a reverse proxy with a custom URL, set `GF_SERVER_ROOT_URL` to match one of the above or add `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=litmus-edge-datasource` to your environment.

## Versioning

Image tags match plugin release versions (e.g., `1.0.0`). The `:latest` tag always points to the most recent release.

To pin a specific Grafana base image version, use the `GRAFANA_VERSION` build argument:

```bash
docker build --build-arg GRAFANA_VERSION=12.4.1 -f litmus-grafana/Dockerfile .
```

## Platform

`linux/amd64` only.
