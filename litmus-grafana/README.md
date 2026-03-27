# litmus-grafana

A preconfigured Grafana image bundled with the Litmus Edge datasource and a curated set of visualization plugins. Deploy a single container to start streaming live edge data into dashboards — no manual plugin installation or datasource setup required.

## Quick start

Run the latest prebuilt image:

```bash
docker run -p 3000:3000 \
  us-docker.pkg.dev/litmus-customer-facing/litmus-solutions/litmus-grafana:latest
```

Open `http://localhost:3000` in your browser (default credentials: `admin` / `admin`).

The Litmus Edge datasource is already installed and provisioned as the default.

## What's included

- **[Litmus Edge datasource](https://github.com/litmusautomation/edge-datasource/blob/main/README.md)** — live data streaming from Litmus Edge via NATS, privately signed
- **[HTML Graphics](https://grafana.com/grafana/plugins/gapit-htmlgraphics-panel/)** — custom HTML/SVG visualizations
<!-- - **[Infinity](https://grafana.com/grafana/plugins/yesoreyeram-infinity-datasource/)** — query any REST, GraphQL, or CSV endpoint -->
- **[Business Variable](https://grafana.com/grafana/plugins/volkovlabs-variable-panel/)** — enhanced dashboard variable controls
- **[Business Input](https://grafana.com/grafana/plugins/marcusolsson-static-datasource/)** — static and mock data for prototyping dashboards
- **[Plotly.js](https://grafana.com/grafana/plugins/nline-plotlyjs-panel/)** — interactive Plotly.js charts

## Usage

### Inside Litmus Edge (default)

When running as a container inside Litmus Edge, the plugin reaches Litmus Edge through `EDGE_DOCKER_GATEWAY_IP` and connects to NATS without credentials. The image does not auto-detect this address, so update it if your instance does not use the default `10.30.50.1` gateway. `EDGE_API_TOKEN` is optional, but recommended for topic discovery:

```bash
docker run -p 3000:3000 \
  -e EDGE_DOCKER_GATEWAY_IP=10.30.50.1 \
  -e EDGE_API_TOKEN=<your-edge-token> \
  us-docker.pkg.dev/litmus-customer-facing/litmus-solutions/litmus-grafana
```

### External Litmus Edge

To connect to a remote Litmus Edge instance, set `EDGE_EXTERNAL=true` and provide the Litmus Edge address, NATS Proxy port, and Access Account token:

```bash
docker run -p 3000:3000 \
  -e EDGE_EXTERNAL=true \
  -e EDGE_HOSTNAME=172.17.0.1:8443 \
  -e EDGE_NATS_PROXY_PORT=4222 \
  -e EDGE_ACCESS_ACCOUNT_TOKEN=<your-access-account-token> \
  -e EDGE_API_TOKEN=<your-edge-token> \
  us-docker.pkg.dev/litmus-customer-facing/litmus-solutions/litmus-grafana
```

The Litmus Edge datasource is automatically provisioned as the default. No manual configuration is needed — just set the environment variables and go.

## Environment variables

| Variable                    | Required      | Description                                                                                                                |
| --------------------------- | ------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `EDGE_EXTERNAL`             | No            | Set to `true` to connect to a remote Litmus Edge instance. Default: `false`                                                |
| `EDGE_DOCKER_GATEWAY_IP`    | No            | Docker gateway IP used inside Litmus Edge. Default: `10.30.50.1`. Update it when your instance uses a different gateway IP |
| `EDGE_HOSTNAME`             | External only | Litmus Edge address. Use `host` or `host:port`                                                                             |
| `EDGE_NATS_PROXY_PORT`      | No            | NATS Proxy port used for live data streaming. Default: `4222`                                                              |
| `EDGE_ACCESS_ACCOUNT_TOKEN` | External only | Access Account token with NATS Proxy read access                                                                           |
| `EDGE_API_TOKEN`            | No            | Optional API token used for topic discovery via the DeviceHub API                                                          |

## Plugin signature

The Litmus Edge datasource is **privately signed** for localhost. The signature is valid when Grafana's `root_url` matches one of:

- `http://localhost:3000` (default)
- `http://localhost:3001`
- `http://localhost:3002`
- `http://localhost:8080`
- `http://localhost:8443`

If Grafana runs behind a reverse proxy with a different URL, either set `GF_SERVER_ROOT_URL` to one of the above or allow the plugin explicitly with `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=litmus-edge-datasource`.

## Using the plugin without this image

If you already have a Grafana deployment, you can install the plugin separately instead of using `litmus-grafana`. See [README.md](https://github.com/litmusautomation/edge-datasource/blob/main/README.md) for Grafana CLI installation and datasource provisioning examples.

## Versioning

Image tags correspond to plugin release versions (e.g., `1.0.0`). The `:latest` tag always points to the most recent release.

To pin a specific Grafana base image version, pass the `GRAFANA_VERSION` build argument:

```bash
docker build --build-arg GRAFANA_VERSION=12.4.1 -f litmus-grafana/Dockerfile .
```

## Platform

`linux/amd64` only.
