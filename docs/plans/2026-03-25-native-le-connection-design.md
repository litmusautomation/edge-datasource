# Native Litmus Edge Connection Mode

**Date:** 2026-03-25
**Status:** Approved
**Complexity:** Medium

## Problem

The Grafana datasource plugin currently requires users to manually configure a hostname and Access Account Token to connect to a Litmus Edge NATS server. When the `litmus-grafana` Docker image runs **inside** Litmus Edge (via the marketplace/containers feature), this is unnecessary — the NATS proxy (`loopedge-access` on port 4222) is reachable from the Docker bridge network without authentication, thanks to the `docker0` whitelist.

## Goal

Make "running inside LE" the default mode. Only the API token (for DeviceHub autocomplete) should be needed. Add a toggle to switch to "external LE" mode for connecting to a remote Litmus Edge instance.

## Investigation Findings

Verified on a live LE 4.0.x instance with a litmus-grafana container running on the default bridge network:

| Aspect | Finding |
|--------|---------|
| Docker network | Default bridge; `docker0` interface IP = container gateway |
| Gateway IP | Detectable from `/proc/net/route` (hex-encoded) |
| Port 4222 | `loopedge-access` — NATS proxy with TLS (self-signed certs) |
| Port 4999 | Raw `nats-server` — not accessible from container |
| Auth (4222) | Whitelist: `lo,lo0,docker0` — no credentials needed from Docker bridge |
| TLS | Self-signed certs; plugin uses `InsecureSkipVerify: true` |
| DeviceHub API | Requires API token (returns 401 without it), reachable at `https://gateway/devicehub/v2` |
| Container → host | Reachable via gateway IP (confirmed with ping, nc, wget) |

## Design

### 1. Settings Model

**TypeScript (`types.ts`):**

```typescript
export interface EdgeDataSourceOptions extends DataSourceJsonData {
  hostname: string;        // only used when externalEdge is true
  externalEdge?: boolean;  // false (default) = inside LE, true = external
}

export interface EdgeSecureJsonData {
  token?: string;     // NATS proxy token — only for external mode
  apiToken?: string;  // DeviceHub API token — optional, both modes
}
```

**Go (`edge/client.go`):**

```go
type ConnectionOptions struct {
    Hostname     string `json:"hostname"`
    Token        string `json:"token"`
    ExternalEdge bool   `json:"externalEdge"`
}
```

### 2. Frontend — ConfigEditor

Add an "External Litmus Edge" toggle (default: OFF), following the same pattern as the existing "Topic autocomplete" toggle.

**Inside LE mode** (`externalEdge: false` — default):

```
┌─ Connection ──────────────────────────────────────┐
│                                                    │
│  ○ External Litmus Edge    [toggle: OFF]           │
│                                                    │
│  (hostname and token fields hidden)                │
│                                                    │
├─ Topic Autocomplete ──────────────────────────────┤
│                                                    │
│  ○ Topic autocomplete      [toggle: ON/OFF]        │
│  API Token: [••••••••••••]  (shown when enabled)   │
│                                                    │
└────────────────────────────────────────────────────┘
```

**External LE mode** (`externalEdge: true`):

```
┌─ Connection ──────────────────────────────────────┐
│                                                    │
│  ○ External Litmus Edge    [toggle: ON]            │
│                                                    │
│  Hostname: [172.16.0.100    ]  (required)          │
│  Access Account Token: [••••••••]  (required)      │
│                                                    │
├─ Topic Autocomplete ──────────────────────────────┤
│                                                    │
│  ○ Topic autocomplete      [toggle: ON/OFF]        │
│  API Token: [••••••••••••]  (shown when enabled)   │
│                                                    │
└────────────────────────────────────────────────────┘
```

**Behavior:**
- When toggling External OFF → clear hostname from `jsonData`
- Keep token in secure storage to avoid re-entry if user toggles back
- Validation: when `externalEdge` is true, hostname and token are required

### 3. Backend — Gateway Detection

New function to auto-resolve the host when running inside LE:

```go
// resolveGatewayHost reads /proc/net/route to find the default gateway IP.
// On the Docker bridge network, this is the host machine running LE.
func resolveGatewayHost() (string, error) {
    // Read /proc/net/route
    // Find line where Destination == "00000000" (default route)
    // Decode hex Gateway field (little-endian) to dotted IP
    // e.g., "01321E0A" → 10.30.50.1
}
```

Fallback: if `/proc/net/route` is unavailable or parsing fails, return a clear error asking the user to switch to external mode.

### 4. Backend — Settings Validation

In `getSettings()`:

```go
if opts.ExternalEdge {
    // External mode: hostname and token are required
    if opts.Hostname == "" {
        return nil, "", fmt.Errorf("hostname is required in external mode")
    }
    if opts.Token == "" {
        return nil, "", fmt.Errorf("Access Account token is required in external mode")
    }
} else {
    // Inside LE: resolve host, no token needed
    gateway, err := resolveGatewayHost()
    if err != nil {
        return nil, "", fmt.Errorf("could not detect gateway host: %w (switch to External mode and provide hostname manually)", err)
    }
    opts.Hostname = gateway
    // opts.Token remains empty — no auth needed from docker0
}
```

### 5. Backend — NATS Connection

Modify `NewClient()` to handle both modes:

```go
natsOpts := []nats.Option{
    nats.Secure(&tls.Config{InsecureSkipVerify: true}),
    // ... reconnection handlers, etc.
}

if opts.Token != "" {
    // External mode: authenticate via user/password
    natsURL = fmt.Sprintf("nats://admin:%s@%s:4222", opts.Token, host)
} else {
    // Inside LE: no auth, whitelisted on docker0
    natsURL = fmt.Sprintf("nats://%s:4222", host)
}
```

### 6. Backend — DeviceHub Client

No change to the client logic. The hostname is resolved the same way (gateway or explicit). API token is always required for autocomplete — the DeviceHub API returns 401 without it.

```go
// Same as today, but hostname may come from gateway detection
endpoint := fmt.Sprintf("https://%s/devicehub/v2", stripPort(hostname))
```

### 7. Provisioning (datasources.yml)

```yaml
apiVersion: 1

datasources:
  - name: 'Litmus Edge'
    type: 'litmus-edge-datasource'
    access: proxy
    isDefault: true
    orgId: 1
    version: 1
    editable: true
    jsonData:
      externalEdge: ${LITMUS_EDGE_EXTERNAL:-false}
      hostname: ${LITMUS_EDGE_HOSTNAME:-}
    secureJsonData:
      token: ${LITMUS_EDGE_ACCESS_ACCOUNT_TOKEN:-}
      apiToken: ${LITMUS_EDGE_API_TOKEN:-}
```

For inside-LE deployments, only `LITMUS_EDGE_API_TOKEN` needs to be set.

### 8. Health Check

Adapt `CheckHealth()`:

- Inside LE: validate NATS connection (no token check); validate DeviceHub if API token is configured
- External: validate NATS connection + token; validate DeviceHub if API token is configured

## Files to Modify

| File | Change |
|------|--------|
| `src/types.ts` | Add `externalEdge` to `EdgeDataSourceOptions` |
| `src/components/ConfigEditor.tsx` | Add "External Litmus Edge" toggle, conditionally show hostname/token |
| `pkg/edge/client.go` | Add `ExternalEdge` field, support no-auth NATS URL, add `resolveGatewayHost()` |
| `pkg/plugin/datasource.go` | Update `getSettings()` validation for both modes |
| `pkg/plugin/health.go` | Adapt health check for inside-LE mode |
| `litmus-grafana/provisioning/datasources/datasources.yml` | Update for new env vars and mode |
