# litmus-grafana Docker Image

## Problem

We decided not to publish the plugin to the Grafana marketplace. Instead, we privately sign the plugin and distribute it via a Docker image that customers can run directly.

## Decision

Ship a `litmus-grafana` Docker image that bundles:

- Grafana (open-source edition, version controllable)
- The `litmus-edge-datasource` plugin, privately signed for localhost
- Community plugins useful for industrial/IoT dashboards
- Provisioned Litmus Edge datasource (configured via env vars at runtime)

The image lives in this repo under `litmus-grafana/` and is pushed to `us-docker.pkg.dev/litmus-customer-facing/litmus-solutions/litmus-grafana`.

## Architecture

### Directory structure

```
litmus-grafana/
├── Dockerfile
└── provisioning/
    └── datasources/
        └── datasources.yml
```

### Dockerfile

```dockerfile
ARG GRAFANA_VERSION=latest
FROM grafana/grafana:${GRAFANA_VERSION}

# Community plugins
RUN grafana-cli plugins install gapit-htmlgraphics-panel && \
    grafana-cli plugins install yesoreyeram-infinity-datasource && \
    grafana-cli plugins install volkovlabs-variable-panel && \
    grafana-cli plugins install marcusolsson-static-datasource && \
    grafana-cli plugins install nline-plotlyjs-panel

# Provisioned datasource (configured via env vars at runtime)
COPY --chown=grafana:root litmus-grafana/provisioning/ /etc/grafana/provisioning/

# Signed plugin (private signature, localhost rootUrls)
COPY --chown=grafana:root dist/ /var/lib/grafana/plugins/litmus-edge-datasource/

ENV GF_SERVER_ROOT_URL=http://localhost:3000
```

Key properties:

- Runs as `grafana` user (image default) — no root
- No dev tools, no supervisor, no livereload
- `GF_SERVER_ROOT_URL` matches the private signature rootUrls
- `GRAFANA_VERSION` pinned in Dockerfile, overridable via workflow_dispatch

### Private signing

The plugin is signed with `--rootUrls` covering common localhost ports:

```
http://localhost:3000
http://localhost:3001
http://localhost:3002
http://localhost:8080
http://localhost:8443
```

Grafana validates the signature against its `root_url` setting. The default is `http://localhost:3000`, so the plugin loads as signed out of the box. If a customer overrides `GF_SERVER_ROOT_URL` to something not in this list, the plugin will be treated as unsigned.

### CI/CD — single workflow, three jobs

**File:** `.github/workflows/release.yml` (replaces existing)

**Triggers:**

- `on: push: tags: ['v*']` — every release
- `on: workflow_dispatch` — manual rebuild (e.g., bump Grafana version)

**Jobs:**

```
v* tag ──→ [build] ──→ [release]  (GitHub Release with signed ZIP)
                   ──→ [docker]   (build + push image to GCR)

manual ──→ [build] ──→ [docker]   (no release created)
```

#### Job 1: `build`

1. Checkout
2. Setup Node 22 + Go
3. `npm ci && npm run build`
4. `mage build:linux && chmod 0755 dist/gpx_edge_linux_amd64`
5. `npx @grafana/sign-plugin@latest --rootUrls http://localhost:3000,...`
6. Upload `dist/` as workflow artifact

#### Job 2: `release` (tag pushes only)

1. Download `dist/` artifact
2. Package as `litmus-edge-datasource-<version>.zip`
3. Create GitHub Release with the ZIP attached

#### Job 3: `docker`

1. Download `dist/` artifact
2. Login to Artifact Registry via `DOCKER_AUTH_CONFIG` secret
3. `docker build --build-arg GRAFANA_VERSION=... -f litmus-grafana/Dockerfile .`
4. Push `<registry>/litmus-grafana:<version>` + `:latest`

### Secrets (GitHub repo)

| Secret                        | Purpose                         |
| ----------------------------- | ------------------------------- |
| `GRAFANA_ACCESS_POLICY_TOKEN` | Plugin signing (already exists) |
| `DOCKER_AUTH_CONFIG`          | Docker config.json for GCR push |

### Grafana version control

- Default is pinned in the Dockerfile `ARG GRAFANA_VERSION=latest`
- Override via `workflow_dispatch` input for manual rebuilds
- Tag pushes use whatever the Dockerfile specifies

### Community plugins included

| Plugin                            | Purpose                             |
| --------------------------------- | ----------------------------------- |
| `gapit-htmlgraphics-panel`        | Custom HTML/SVG visualizations      |
| `yesoreyeram-infinity-datasource` | Query any REST/GraphQL/CSV endpoint |
| `volkovlabs-variable-panel`       | Enhanced dashboard variable UI      |
| `marcusolsson-static-datasource`  | Static/mock data for dashboards     |
| `nline-plotlyjs-panel`            | Plotly.js charts                    |

**Dropped** (incompatible or deprecated for Grafana 12.x):

- `cloudspout-button-panel` — deprecated
- `isaozler-paretochart-panel` — stale since 2022
- `snuids-svg-panel` — stale, HTML Graphics covers this

### Limitations

- **rootUrls constraint**: private signature only validates against the listed localhost URLs. Customers with custom `root_url` settings would need `allow_loading_unsigned_plugins: litmus-edge-datasource` in their Grafana config.
- **linux/amd64 only**: backend binary is built for linux/amd64.
- **Datasource requires env vars**: customers must set `LITMUS_EDGE_HOSTNAME`, `LITMUS_EDGE_ACCESS_ACCOUNT_TOKEN`, and `LITMUS_EDGE_API_TOKEN` at runtime.

## Changes from current state

| File                            | Change                                                                       |
| ------------------------------- | ---------------------------------------------------------------------------- |
| `.github/workflows/release.yml` | Replace `build-plugin` action with manual build + private sign + docker push |
| `litmus-grafana/Dockerfile`     | New                                                                          |
| `litmus-grafana/provisioning/`  | New (datasource provisioning via env vars)                                   |
