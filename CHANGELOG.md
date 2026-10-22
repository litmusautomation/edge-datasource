# Changelog

## 0.2.0

### Security

- Build with Go toolchain 1.26.10 (fixes stdlib `net/http`, `crypto/tls`, `crypto/x509` CVEs); upgrade `golang.org/x/net` to v0.61.0, `golang.org/x/crypto` to v0.58.0 and `grafana-plugin-sdk-go` to v0.297.0.
- Upgrade `@grafana/*` packages to 12.4.13, run `npm audit fix`, and pin `moment` 2.31.0 and `uuid` ^11.1.1 (clears the critical `protobufjs` / `websocket-driver` alerts).
- Docker image base pinned to Grafana 12.4.13 (latest patched 12.x); Grafana 13 is not yet supported.

## 0.1.1

### Dependencies

- Upgrade OpenTelemetry Go modules to v1.43.0 and related `golang.org/x` / gRPC pins (addresses Dependabot alerts).
- Pin `follow-redirects` to ^1.16.0 and refresh npm lockfile (axios, lodash) for patched transitive versions.

### Configuration

- External mode: `EDGE_ACCESS_ACCOUNT_TOKEN` was renamed to `EDGE_ACCESS_ACCOUNT_API_KEY` — update `.env` and any deployment config that set the old name.

## 0.1.0 — Initial release

### Features

- **Real-time data streaming** — subscribe to NATS topics on Litmus Edge and stream live frames into Grafana panels via Grafana Live (WebSocket). No polling, no historical database required.
- **DeviceHub support** — automatically parse DeviceHub messages and expose metadata labels (`tagName`, `deviceName`, `deviceId`, `datatype`, `description`, `registerId`) for use in legends and transformations.
- **Inside-LE connection mode** — when deployed on Litmus Edge, the plugin auto-discovers the host via the Docker bridge network and connects to NATS without credentials. Zero configuration required.
- **External connection mode** — connect to a remote Litmus Edge instance with a Litmus Edge address, NATS Proxy port, and Access Account API Key.
- **Topic discovery** — use an API Token to query available topics from the Litmus Edge DeviceHub API directly in the query editor.
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
