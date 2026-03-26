# litmus-grafana

A preconfigured Grafana image bundled with the Litmus Edge datasource and a curated set of visualization plugins. Deploy a single container to start streaming live edge data into dashboards — no manual plugin installation or datasource setup required.

## What's included

- **[Litmus Edge datasource](../README.md)** — live data streaming from Litmus Edge via NATS, privately signed
- **[HTML Graphics](https://grafana.com/grafana/plugins/gapit-htmlgraphics-panel/)** — custom HTML/SVG visualizations
<!-- - **[Infinity](https://grafana.com/grafana/plugins/yesoreyeram-infinity-datasource/)** — query any REST, GraphQL, or CSV endpoint -->
- **[Business Variable](https://grafana.com/grafana/plugins/volkovlabs-variable-panel/)** — enhanced dashboard variable controls
- **[Business Input](https://grafana.com/grafana/plugins/marcusolsson-static-datasource/)** — static and mock data for prototyping dashboards
- **[Plotly.js](https://grafana.com/grafana/plugins/nline-plotlyjs-panel/)** — interactive Plotly.js charts

## Usage

### Inside Litmus Edge (default)

When running as a container inside Litmus Edge, the plugin discovers the host automatically via the Docker bridge network and connects to NATS without credentials. Only an API token is needed for topic autocomplete:

```bash
docker run -p 3000:3000 \
  -e LITMUS_EDGE_API_TOKEN=<your-api-token> \
  us-docker.pkg.dev/litmus-customer-facing/litmus-solutions/litmus-grafana
```

### External Litmus Edge

To connect to a remote Litmus Edge instance, enable external mode and provide the hostname and Access Account token:

```bash
docker run -p 3000:3000 \
  -e LITMUS_EDGE_EXTERNAL=true \
  -e LITMUS_EDGE_HOSTNAME=172.17.0.1 \
  -e LITMUS_EDGE_ACCESS_ACCOUNT_TOKEN=<your-access-account-token> \
  -e LITMUS_EDGE_API_TOKEN=<your-api-token> \
  us-docker.pkg.dev/litmus-customer-facing/litmus-solutions/litmus-grafana
```

Open `http://localhost:3000` in your browser (default credentials: `admin` / `admin`).

The Litmus Edge datasource is automatically provisioned as the default. No manual configuration is needed — just set the environment variables and go.

## Environment variables

| Variable | Required | Description |
| --- | --- | --- |
| `LITMUS_EDGE_EXTERNAL` | No | Enable external mode to connect to a remote Litmus Edge instance. Default: `false` |
| `LITMUS_EDGE_HOSTNAME` | External only | Hostname or IP of the Litmus Edge instance |
| `LITMUS_EDGE_ACCESS_ACCOUNT_TOKEN` | External only | Access Account token with NATS Proxy read access |
| `LITMUS_EDGE_API_TOKEN` | No | API token for topic autocomplete via the DeviceHub API |

## Plugin signature

The Litmus Edge datasource is **privately signed** for localhost. The signature is valid when Grafana's `root_url` matches one of:

- `http://localhost:3000` (default)
- `http://localhost:3001`
- `http://localhost:3002`
- `http://localhost:8080`
- `http://localhost:8443`

If Grafana runs behind a reverse proxy with a different URL, either set `GF_SERVER_ROOT_URL` to one of the above or allow the plugin explicitly with `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=litmus-edge-datasource`.

## Versioning

Image tags correspond to plugin release versions (e.g., `1.0.0`). The `:latest` tag always points to the most recent release.

To pin a specific Grafana base image version, pass the `GRAFANA_VERSION` build argument:

```bash
docker build --build-arg GRAFANA_VERSION=12.4.1 -f litmus-grafana/Dockerfile .
```

## Platform

`linux/amd64` only.
