# Changelog

## 0.1.0 — Initial release

### Features

- **Real-time data streaming** — subscribe to NATS topics on Litmus Edge and stream live frames into Grafana panels via Grafana Live (WebSocket). No polling, no historical database required.
- **DeviceHub support** — automatically parse DeviceHub messages and expose metadata labels (`tagName`, `deviceName`, `deviceId`, `datatype`, `description`, `registerId`) for use in legends and transformations.
- **Inside-LE connection mode** — when deployed on Litmus Edge, the plugin auto-discovers the host via the Docker bridge network and connects to NATS without credentials. Zero configuration required.
- **External connection mode** — connect to a remote Litmus Edge instance with hostname and Access Account token.
- **Topic autocomplete** — query available topics from the Litmus Edge DeviceHub API directly in the query editor.
- **Template variable support** — use Grafana variables in topic fields (`$site.$area.$line.$sensor`) for reusable dashboards across sites and production lines.
- **litmus-grafana Docker image** — preconfigured Grafana image with the Litmus Edge datasource and curated visualization plugins (HTML Graphics, Business Variable, Business Input, Plotly.js), provisioned and ready to run.

### Operational

- O(1) NATS message routing with concurrent-safe subscription management and a 10 000-message buffer per topic.
- Automatic NATS reconnection with exponential backoff.
- OpenTelemetry tracing spans for observability.
- Plugin privately signed for localhost (`3000`, `3001`, `3002`, `8080`, `8443`).

### CI/CD

- Grafana reusable CI workflow with Playwright E2E tests.
- Release workflow: build, sign, GitHub Release with build attestation, and Docker image push to Artifact Registry.
- Monthly scaffolding updater via `grafana/plugin-actions/create-plugin-update`.
